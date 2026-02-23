// internal/infrastructure/messaging/sqs_consumer.go
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/farmanexo/price-service/internal/domain/entities"
	"github.com/farmanexo/price-service/internal/domain/events"
	"github.com/farmanexo/price-service/internal/domain/repositories"
	"github.com/farmanexo/price-service/internal/domain/services"
	"github.com/farmanexo/price-service/pkg/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SQSConsumer consume eventos de la cola SQS del Pharmacy Service
type SQSConsumer struct {
	sqsClient        *sqs.Client
	queueURL         string
	priceHistoryRepo repositories.PriceHistoryRepository
	alertRepo        repositories.PriceAlertRepository
	eventPublisher   services.EventPublisher
	logger           *zap.Logger
	pollInterval     time.Duration
	maxMessages      int32
	visibilityTimeout int32
}

// NewSQSConsumer crea una nueva instancia del consumidor SQS
func NewSQSConsumer(
	awsCfg config.AWSConfig,
	sqsCfg config.SQSConfig,
	consumerCfg config.ConsumerConfig,
	priceHistoryRepo repositories.PriceHistoryRepository,
	alertRepo repositories.PriceAlertRepository,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) (*SQSConsumer, error) {
	optFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(awsCfg.Region),
	}

	if awsCfg.Endpoint != "" {
		optFns = append(optFns, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		))
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS para consumer: %w", err)
	}

	sqsOptFns := []func(*sqs.Options){}
	if awsCfg.Endpoint != "" {
		sqsOptFns = append(sqsOptFns, func(o *sqs.Options) {
			o.BaseEndpoint = &awsCfg.Endpoint
		})
	}

	sqsClient := sqs.NewFromConfig(cfg, sqsOptFns...)

	pollInterval := consumerCfg.PollInterval
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}
	maxMessages := consumerCfg.MaxMessages
	if maxMessages == 0 {
		maxMessages = 10
	}
	visibilityTimeout := consumerCfg.VisibilityTimeout
	if visibilityTimeout == 0 {
		visibilityTimeout = 30
	}

	logger.Info("SQS Consumer inicializado",
		zap.String("queue_url", sqsCfg.PharmacyEventsQueueURL),
		zap.Duration("poll_interval", pollInterval),
		zap.Int32("max_messages", maxMessages),
	)

	return &SQSConsumer{
		sqsClient:         sqsClient,
		queueURL:          sqsCfg.PharmacyEventsQueueURL,
		priceHistoryRepo:  priceHistoryRepo,
		alertRepo:         alertRepo,
		eventPublisher:    eventPublisher,
		logger:            logger,
		pollInterval:      pollInterval,
		maxMessages:       maxMessages,
		visibilityTimeout: visibilityTimeout,
	}, nil
}

// Start inicia el consumo de mensajes en un goroutine
func (c *SQSConsumer) Start(ctx context.Context) {
	c.logger.Info("Iniciando consumidor SQS para pharmacy events")

	go func() {
		ticker := time.NewTicker(c.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Deteniendo consumidor SQS")
				return
			case <-ticker.C:
				c.pollMessages(ctx)
			}
		}
	}()
}

func (c *SQSConsumer) pollMessages(ctx context.Context) {
	output, err := c.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &c.queueURL,
		MaxNumberOfMessages: c.maxMessages,
		WaitTimeSeconds:     5,
		VisibilityTimeout:   c.visibilityTimeout,
	})
	if err != nil {
		c.logger.Error("Error recibiendo mensajes de SQS", zap.Error(err))
		return
	}

	for _, msg := range output.Messages {
		c.processMessage(ctx, msg)
	}
}

func (c *SQSConsumer) processMessage(ctx context.Context, msg sqstypes.Message) {
	if msg.Body == nil {
		return
	}

	var event events.InventoryUpdatedEvent
	if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
		c.logger.Error("Error deserializando evento",
			zap.Error(err),
			zap.String("body", *msg.Body),
		)
		c.deleteMessage(ctx, msg)
		return
	}

	if event.EventType != "INVENTORY_UPDATED" {
		c.logger.Debug("Evento ignorado", zap.String("event_type", event.EventType))
		c.deleteMessage(ctx, msg)
		return
	}

	c.logger.Info("Procesando evento INVENTORY_UPDATED",
		zap.String("product_id", event.ProductID),
		zap.String("pharmacy_id", event.PharmacyID),
		zap.Float64("price", event.Price),
	)

	// Buscar precio anterior
	var previousPrice *float64
	latestPrice, err := c.priceHistoryRepo.FindLatestPrice(ctx, event.ProductID, event.PharmacyID)
	if err == nil && latestPrice != nil {
		previousPrice = &latestPrice.Price
	}

	// Registrar nuevo precio
	history := &entities.PriceHistory{
		ID:            uuid.New().String(),
		ProductID:     event.ProductID,
		PharmacyID:    event.PharmacyID,
		PharmacyName:  event.Metadata.PharmacyName,
		ProductName:   event.Metadata.ProductName,
		Price:         event.Price,
		PreviousPrice: previousPrice,
		Currency:      "PEN",
		Source:        "inventory_update",
		RecordedAt:    time.Now(),
	}

	if err := c.priceHistoryRepo.Create(ctx, history); err != nil {
		c.logger.Error("Error registrando precio", zap.Error(err))
		return
	}

	// Verificar alertas activas
	c.checkAlerts(ctx, event.ProductID, event.PharmacyID, event.Price, event.Metadata.PharmacyName)

	// Eliminar mensaje de la cola
	c.deleteMessage(ctx, msg)
}

func (c *SQSConsumer) checkAlerts(ctx context.Context, productID, pharmacyID string, price float64, pharmacyName string) {
	alerts, err := c.alertRepo.FindActiveByProductID(ctx, productID)
	if err != nil {
		c.logger.Error("Error buscando alertas activas", zap.Error(err))
		return
	}

	for _, alert := range alerts {
		if price <= alert.TargetPrice {
			now := time.Now()
			alert.IsTriggered = true
			alert.TriggeredAt = &now
			alert.TriggeredPharmacyID = &pharmacyID
			alert.TriggeredPharmacyName = &pharmacyName
			alert.TriggeredPrice = &price
			alert.CurrentPrice = &price

			if err := c.alertRepo.Update(ctx, &alert); err != nil {
				c.logger.Error("Error actualizando alerta", zap.Error(err))
				continue
			}

			// Publicar evento de alerta disparada
			event := events.NewPriceEvent(events.EventPriceAlertTriggered).
				WithProduct(productID).
				WithPharmacy(pharmacyID).
				WithUser(alert.UserID)
			event.Metadata["alert_id"] = alert.ID
			event.Metadata["target_price"] = fmt.Sprintf("%.2f", alert.TargetPrice)
			event.Metadata["triggered_price"] = fmt.Sprintf("%.2f", price)

			go func(evt events.PriceEvent) {
				if err := c.eventPublisher.Publish(context.Background(), evt); err != nil {
					c.logger.Error("Error publicando evento de alerta", zap.Error(err))
				}
			}(event)

			c.logger.Info("Alerta de precio disparada",
				zap.String("alert_id", alert.ID),
				zap.String("user_id", alert.UserID),
				zap.Float64("target_price", alert.TargetPrice),
				zap.Float64("actual_price", price),
			)
		} else {
			// Actualizar precio actual de la alerta
			alert.CurrentPrice = &price
			if err := c.alertRepo.Update(ctx, &alert); err != nil {
				c.logger.Error("Error actualizando precio actual de alerta", zap.Error(err))
			}
		}
	}
}

func (c *SQSConsumer) deleteMessage(ctx context.Context, msg sqstypes.Message) {
	_, err := c.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &c.queueURL,
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil {
		c.logger.Error("Error eliminando mensaje de SQS", zap.Error(err))
	}
}
