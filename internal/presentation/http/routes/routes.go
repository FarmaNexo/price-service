// internal/presentation/http/routes/routes.go
package routes

import (
	"net/http"

	"github.com/farmanexo/price-service/internal/presentation/http/controllers"
	"github.com/farmanexo/price-service/internal/presentation/http/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// SetupRoutes configura todas las rutas del servicio
func SetupRoutes(
	priceController *controllers.PriceController,
	authMiddleware *middlewares.AuthMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	// ========================================
	// MIDDLEWARES GLOBALES
	// ========================================

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://farmanexo.pe"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middlewares.CorrelationID)

	// ========================================
	// SWAGGER DOCUMENTATION
	// ========================================

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:4005/swagger/doc.json"),
	))

	// ========================================
	// HEALTH CHECK
	// ========================================

	r.Get("/health", priceController.HealthCheck)
	r.Get("/prices/health", priceController.HealthCheck)
	r.Get("/", priceController.HealthCheck)

	// ========================================
	// API ROUTES - VERSION 1
	// ========================================

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/prices", func(r chi.Router) {
			// ========================================
			// ENDPOINTS PÚBLICOS
			// ========================================
			r.Post("/compare", priceController.ComparePrices)
			r.Get("/compare/generic-vs-brand/{id}", priceController.GetGenericVsBrand)
			r.Get("/history/{id}", priceController.GetPriceHistory)
			// HU-015 — Alternativas terapéuticas en tiempo real (mismo DCI, ordenadas por ahorro)
			r.Get("/products/{id}/alternatives", priceController.GetProductAlternatives)

			// ========================================
			// ENDPOINTS AUTENTICADOS (Requiere JWT)
			// ========================================
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)

				r.Post("/alerts", priceController.CreatePriceAlert)
				r.Get("/alerts", priceController.ListPriceAlerts)
				r.Delete("/alerts/{id}", priceController.DeletePriceAlert)
			})

			// ========================================
			// ENDPOINTS ADMIN (Requiere JWT + Admin)
			// ========================================
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Use(authMiddleware.RequireAdmin)

				r.Get("/stats/{id}", priceController.GetPriceStats)
				r.Post("/record", priceController.RecordPrice)
			})
		})
	})

	// ========================================
	// API ROUTES - VERSION 2 (Futuro)
	// ========================================

	r.Route("/api/v2", func(r chi.Router) {
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("API v2 - Próximamente"))
		})
	})

	return r
}
