// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/farmanexo/price-service/internal/application/commands"
	"github.com/farmanexo/price-service/internal/application/handlers"
	"github.com/farmanexo/price-service/internal/application/postprocessors"
	"github.com/farmanexo/price-service/internal/application/preprocessors"
	"github.com/farmanexo/price-service/internal/application/validators"
	"github.com/farmanexo/price-service/internal/presentation/dto/responses"
	"github.com/farmanexo/price-service/internal/infrastructure/cache"
	"github.com/farmanexo/price-service/internal/infrastructure/clients"
	"github.com/farmanexo/price-service/internal/infrastructure/messaging"
	"github.com/farmanexo/price-service/internal/infrastructure/persistence/postgres"
	"github.com/farmanexo/price-service/internal/infrastructure/security"
	"github.com/farmanexo/price-service/internal/presentation/http/controllers"
	"github.com/farmanexo/price-service/internal/presentation/http/middlewares"
	"github.com/farmanexo/price-service/internal/presentation/http/routes"
	"github.com/farmanexo/price-service/pkg/config"
	"github.com/farmanexo/price-service/pkg/logger"
	"github.com/farmanexo/price-service/pkg/mediator"

	// Swagger docs
	_ "github.com/farmanexo/price-service/docs"

	"go.uber.org/zap"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// @title           FarmaNexo Price Service API
// @version         1.0
// @description     Servicio de comparación y seguimiento de precios farmacéuticos para FarmaNexo - Microservicio con CQRS y Clean Architecture
// @termsOfService  https://farmanexo.pe/terms

// @contact.name    FarmaNexo API Support
// @contact.url     https://farmanexo.pe/support
// @contact.email   support@farmanexo.pe

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:4005
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

// @tag.name         Prices
// @tag.description  Endpoints de comparación y consulta de precios

// @tag.name         Alerts
// @tag.description  Endpoints de alertas de precio

// @tag.name         Admin
// @tag.description  Endpoints de administración de precios

// @tag.name         Health
// @tag.description  Endpoints de salud del servicio

func main() {
	env := getEnvironment()
	cfg, err := config.LoadConfig(env)
	if err != nil {
		panic(fmt.Sprintf("Error cargando configuración: %v", err))
	}

	zapLogger, err := logger.NewLogger(cfg.Environment, cfg.Log.Encoding, cfg.Log.Level)
	if err != nil {
		panic(fmt.Sprintf("Error inicializando logger: %v", err))
	}
	defer zapLogger.Sync()

	zapLogger.Info("Iniciando Price Service",
		zap.String("environment", cfg.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	db := initDatabase(cfg, zapLogger)

	zapLogger.Info("Auto-migration deshabilitado - Usar migraciones manuales")

	// ========================================
	// REPOSITORIOS
	// ========================================
	priceHistoryRepo := postgres.NewPriceHistoryRepository(db, zapLogger)
	alertRepo := postgres.NewPriceAlertRepository(db, zapLogger)
	genericBrandRepo := postgres.NewGenericBrandRepository(db, zapLogger)

	// ========================================
	// SERVICIOS
	// ========================================
	jwtService := security.NewJWTService(cfg.JWT.Secret, zapLogger)

	// SQS Event Publisher
	eventPublisher, err := messaging.NewSQSEventPublisher(cfg.AWS, cfg.SQS, zapLogger)
	if err != nil {
		zapLogger.Fatal("Error inicializando SQS EventPublisher", zap.Error(err))
	}

	// Redis Cache
	redisClient, err := cache.NewRedisClient(cfg.Redis, cfg.Environment, zapLogger)
	if err != nil {
		zapLogger.Fatal("Error inicializando Redis", zap.Error(err))
	}
	defer redisClient.Close()

	cacheService := cache.NewRedisCacheService(redisClient, zapLogger)

	// HTTP Clients
	pharmacyClient := clients.NewPharmacyClient(cfg.Services.PharmacyService.BaseURL, zapLogger)
	_ = clients.NewCatalogClient(cfg.Services.CatalogService.BaseURL, zapLogger)

	// ========================================
	// SQS CONSUMER
	// ========================================
	if cfg.Consumer.Enabled {
		sqsConsumer, err := messaging.NewSQSConsumer(
			cfg.AWS, cfg.SQS, cfg.Consumer,
			priceHistoryRepo, alertRepo, eventPublisher,
			zapLogger,
		)
		if err != nil {
			zapLogger.Fatal("Error inicializando SQS Consumer", zap.Error(err))
		}

		consumerCtx, consumerCancel := context.WithCancel(context.Background())
		defer consumerCancel()
		sqsConsumer.Start(consumerCtx)
		zapLogger.Info("SQS Consumer iniciado para pharmacy events")
	}

	// ========================================
	// MEDIATOR
	// ========================================
	med := mediator.NewMediator()

	// ========================================
	// HANDLERS
	// ========================================
	comparePricesHandler := handlers.NewComparePricesHandler(priceHistoryRepo, pharmacyClient, cacheService, zapLogger)
	mediator.RegisterHandler(med, comparePricesHandler)

	getPriceHistoryHandler := handlers.NewGetPriceHistoryHandler(priceHistoryRepo, cacheService, zapLogger)
	mediator.RegisterHandler(med, getPriceHistoryHandler)

	createPriceAlertHandler := handlers.NewCreatePriceAlertHandler(alertRepo, cacheService, zapLogger)
	mediator.RegisterHandler(med, createPriceAlertHandler)

	deletePriceAlertHandler := handlers.NewDeletePriceAlertHandler(alertRepo, cacheService, zapLogger)
	mediator.RegisterHandler(med, deletePriceAlertHandler)

	listPriceAlertsHandler := handlers.NewListPriceAlertsHandler(alertRepo, cacheService, zapLogger)
	mediator.RegisterHandler(med, listPriceAlertsHandler)

	getGenericVsBrandHandler := handlers.NewGetGenericVsBrandHandler(genericBrandRepo, cacheService, zapLogger)
	mediator.RegisterHandler(med, getGenericVsBrandHandler)

	getPriceStatsHandler := handlers.NewGetPriceStatsHandler(priceHistoryRepo, cacheService, zapLogger)
	mediator.RegisterHandler(med, getPriceStatsHandler)

	recordPriceHandler := handlers.NewRecordPriceHandler(priceHistoryRepo, cacheService, zapLogger)
	mediator.RegisterHandler(med, recordPriceHandler)

	// ========================================
	// VALIDATORS
	// ========================================
	comparePricesValidator := validators.NewComparePricesValidator()
	mediator.RegisterValidator[commands.ComparePricesCommand, responses.PriceComparisonResponse](med, comparePricesValidator)

	createPriceAlertValidator := validators.NewCreatePriceAlertValidator()
	mediator.RegisterValidator[commands.CreatePriceAlertCommand, responses.PriceAlertResponse](med, createPriceAlertValidator)

	// ========================================
	// PREPROCESSORS Y POSTPROCESSORS
	// ========================================
	sanitizePreProcessor := preprocessors.NewSanitizeInputPreProcessor(zapLogger)
	med.RegisterPreProcessor(sanitizePreProcessor)

	auditPostProcessor := postprocessors.NewLogAuditPostProcessor(zapLogger)
	med.RegisterPostProcessor(auditPostProcessor)

	zapLogger.Info("Mediator configurado",
		zap.Int("handlers", 8),
		zap.Int("validators", 2),
		zap.Int("preprocessors", 1),
		zap.Int("postprocessors", 1),
	)

	// ========================================
	// MIDDLEWARES
	// ========================================
	authMiddleware := middlewares.NewAuthMiddleware(jwtService, zapLogger)

	// ========================================
	// CONTROLADORES Y RUTAS
	// ========================================
	priceController := controllers.NewPriceController(med, zapLogger)
	router := routes.SetupRoutes(priceController, authMiddleware)

	// ========================================
	// SERVIDOR HTTP
	// ========================================
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		zapLogger.Info("Servidor HTTP iniciado",
			zap.String("address", server.Addr),
			zap.String("swagger_url", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Server.Port)),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Error iniciando servidor", zap.Error(err))
		}
	}()

	// ========================================
	// GRACEFUL SHUTDOWN
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Iniciando graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		zapLogger.Error("Error en shutdown", zap.Error(err))
	}

	zapLogger.Info("Servidor detenido exitosamente")
}

func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	return env
}

func initDatabase(cfg *config.Config, log *zap.Logger) *gorm.DB {
	gormLogLevel := gormlogger.Silent
	if cfg.IsDevelopment() {
		gormLogLevel = gormlogger.Info
	}

	gormLogger := gormlogger.Default.LogMode(gormLogLevel)

	db, err := gorm.Open(pgdriver.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		log.Fatal("Error conectando a PostgreSQL",
			zap.Error(err),
			zap.String("host", cfg.Database.Host),
			zap.Int("port", cfg.Database.Port),
		)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Error obteniendo SQL DB", zap.Error(err))
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	log.Info("Conexión a PostgreSQL establecida",
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.DBName),
	)

	return db
}
