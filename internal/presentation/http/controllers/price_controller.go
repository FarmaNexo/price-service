// internal/presentation/http/controllers/price_controller.go
package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/farmanexo/price-service/internal/application/commands"
	"github.com/farmanexo/price-service/internal/application/queries"
	"github.com/farmanexo/price-service/internal/presentation/dto/requests"
	"github.com/farmanexo/price-service/internal/presentation/dto/responses"
	"github.com/farmanexo/price-service/internal/presentation/http/middlewares"
	"github.com/farmanexo/price-service/internal/shared/common"
	"github.com/farmanexo/price-service/pkg/mediator"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// PriceController controlador HTTP de precios
type PriceController struct {
	mediator *mediator.Mediator
	logger   *zap.Logger
}

func NewPriceController(med *mediator.Mediator, logger *zap.Logger) *PriceController {
	return &PriceController{mediator: med, logger: logger}
}

// respondJSON helper para escribir respuesta JSON
func (c *PriceController) respondJSON(w http.ResponseWriter, response interface{}) {
	statusCode := http.StatusOK

	if resp, ok := response.(interface{ GetHttpStatus() *int }); ok {
		if httpStatus := resp.GetHttpStatus(); httpStatus != nil {
			statusCode = *httpStatus
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		c.logger.Error("Error codificando respuesta JSON", zap.Error(err))
	}
}

// HealthCheck godoc
// @Summary      Health check del servicio
// @Description  Retorna el estado del servicio
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Servicio saludable"
// @Router       /health [get]
func (c *PriceController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type HealthResponse struct {
		Status  string `json:"status" example:"healthy"`
		Service string `json:"service" example:"price-service"`
		Version string `json:"version" example:"1.0.0"`
	}

	health := HealthResponse{
		Status:  "healthy",
		Service: "price-service",
		Version: "1.0.0",
	}

	c.respondJSON(w, common.OkResponse(health))
}

// ========================================
// COMPARACIÓN DE PRECIOS
// ========================================

// ComparePrices godoc
// @Summary      Comparar precios de un producto
// @Description  Compara precios de un producto entre diferentes farmacias
// @Tags         Prices
// @Accept       json
// @Produce      json
// @Param        body  body      requests.ComparePricesRequest  true  "Datos de comparación"
// @Success      200   {object}  common.ApiResponse[responses.PriceComparisonResponse]
// @Failure      400   {object}  common.ApiResponse[responses.PriceComparisonResponse]
// @Failure      500   {object}  common.ApiResponse[responses.PriceComparisonResponse]
// @Router       /api/v1/prices/compare [post]
func (c *PriceController) ComparePrices(w http.ResponseWriter, r *http.Request) {
	var req requests.ComparePricesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.PriceComparisonResponse](
			"VAL_001", "Error en formato de datos: "+err.Error(),
		))
		return
	}

	cmd := commands.ComparePricesCommand{
		ProductID: req.ProductID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		RadiusKm:  req.RadiusKm,
	}

	response, _ := mediator.Send[commands.ComparePricesCommand, responses.PriceComparisonResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// ========================================
// HISTORIAL DE PRECIOS
// ========================================

// GetPriceHistory godoc
// @Summary      Obtener historial de precios
// @Description  Retorna el historial de precios de un producto
// @Tags         Prices
// @Accept       json
// @Produce      json
// @Param        id          path     string  true   "Product ID"
// @Param        pharmacy_id query    string  false  "Pharmacy ID"
// @Param        limit       query    int     false  "Límite de registros"  default(50)
// @Success      200  {object}  common.ApiResponse[responses.PriceHistoryListResponse]
// @Failure      500  {object}  common.ApiResponse[responses.PriceHistoryListResponse]
// @Router       /api/v1/prices/history/{id} [get]
func (c *PriceController) GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	pharmacyID := r.URL.Query().Get("pharmacy_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	query := queries.GetPriceHistoryQuery{
		ProductID:  productID,
		PharmacyID: pharmacyID,
		Limit:      limit,
	}

	response, _ := mediator.Send[queries.GetPriceHistoryQuery, responses.PriceHistoryListResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// ========================================
// GENÉRICO VS MARCA
// ========================================

// GetGenericVsBrand godoc
// @Summary      Comparación genérico vs marca
// @Description  Retorna comparación entre genérico y marca para un producto
// @Tags         Prices
// @Produce      json
// @Param        id  path     string  true  "Product ID"
// @Success      200  {object}  common.ApiResponse[responses.GenericVsBrandResponse]
// @Failure      500  {object}  common.ApiResponse[responses.GenericVsBrandResponse]
// @Router       /api/v1/prices/compare/generic-vs-brand/{id} [get]
func (c *PriceController) GetGenericVsBrand(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")

	query := queries.GetGenericVsBrandQuery{
		ProductID: productID,
	}

	response, _ := mediator.Send[queries.GetGenericVsBrandQuery, responses.GenericVsBrandResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// ========================================
// ALERTAS DE PRECIO
// ========================================

// CreatePriceAlert godoc
// @Summary      Crear alerta de precio
// @Description  Crea una alerta de precio para un producto
// @Tags         Alerts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      requests.CreatePriceAlertRequest  true  "Datos de la alerta"
// @Success      201   {object}  common.ApiResponse[responses.PriceAlertResponse]
// @Failure      400   {object}  common.ApiResponse[responses.PriceAlertResponse]
// @Failure      401   {object}  common.ApiResponse[responses.PriceAlertResponse]
// @Failure      409   {object}  common.ApiResponse[responses.PriceAlertResponse]
// @Router       /api/v1/prices/alerts [post]
func (c *PriceController) CreatePriceAlert(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.PriceAlertResponse]("Usuario no autenticado"))
		return
	}

	var req requests.CreatePriceAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.PriceAlertResponse](
			"VAL_001", "Error en formato de datos: "+err.Error(),
		))
		return
	}

	cmd := commands.CreatePriceAlertCommand{
		UserID:      userID,
		ProductID:   req.ProductID,
		ProductName: req.ProductName,
		TargetPrice: req.TargetPrice,
	}

	response, _ := mediator.Send[commands.CreatePriceAlertCommand, responses.PriceAlertResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// ListPriceAlerts godoc
// @Summary      Listar alertas de precio
// @Description  Retorna las alertas de precio del usuario autenticado
// @Tags         Alerts
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.PriceAlertListResponse]
// @Failure      401  {object}  common.ApiResponse[responses.PriceAlertListResponse]
// @Router       /api/v1/prices/alerts [get]
func (c *PriceController) ListPriceAlerts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.PriceAlertListResponse]("Usuario no autenticado"))
		return
	}

	query := queries.ListPriceAlertsQuery{
		UserID: userID,
	}

	response, _ := mediator.Send[queries.ListPriceAlertsQuery, responses.PriceAlertListResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// DeletePriceAlert godoc
// @Summary      Eliminar alerta de precio
// @Description  Elimina una alerta de precio del usuario autenticado
// @Tags         Alerts
// @Produce      json
// @Security     BearerAuth
// @Param        id  path     string  true  "Alert ID"
// @Success      200  {object}  common.ApiResponse[responses.DeleteAlertResponse]
// @Failure      401  {object}  common.ApiResponse[responses.DeleteAlertResponse]
// @Failure      404  {object}  common.ApiResponse[responses.DeleteAlertResponse]
// @Router       /api/v1/prices/alerts/{id} [delete]
func (c *PriceController) DeletePriceAlert(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.DeleteAlertResponse]("Usuario no autenticado"))
		return
	}

	alertID := chi.URLParam(r, "id")

	cmd := commands.DeletePriceAlertCommand{
		AlertID: alertID,
		UserID:  userID,
	}

	response, _ := mediator.Send[commands.DeletePriceAlertCommand, responses.DeleteAlertResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// ========================================
// ESTADÍSTICAS (ADMIN)
// ========================================

// GetPriceStats godoc
// @Summary      Estadísticas de precios (Admin)
// @Description  Retorna estadísticas de precios de un producto
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id  path     string  true  "Product ID"
// @Success      200  {object}  common.ApiResponse[responses.PriceStatsResponse]
// @Failure      401  {object}  common.ApiResponse[responses.PriceStatsResponse]
// @Failure      403  {object}  common.ApiResponse[responses.PriceStatsResponse]
// @Router       /api/v1/prices/stats/{id} [get]
func (c *PriceController) GetPriceStats(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")

	query := queries.GetPriceStatsQuery{
		ProductID: productID,
	}

	response, _ := mediator.Send[queries.GetPriceStatsQuery, responses.PriceStatsResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// RecordPrice godoc
// @Summary      Registrar precio manualmente (Admin)
// @Description  Registra un precio de producto en una farmacia
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      requests.RecordPriceRequest  true  "Datos del precio"
// @Success      201   {object}  common.ApiResponse[responses.PriceHistoryItem]
// @Failure      400   {object}  common.ApiResponse[responses.PriceHistoryItem]
// @Failure      401   {object}  common.ApiResponse[responses.PriceHistoryItem]
// @Failure      403   {object}  common.ApiResponse[responses.PriceHistoryItem]
// @Router       /api/v1/prices/record [post]
func (c *PriceController) RecordPrice(w http.ResponseWriter, r *http.Request) {
	var req requests.RecordPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.PriceHistoryItem](
			"VAL_001", "Error en formato de datos: "+err.Error(),
		))
		return
	}

	cmd := commands.RecordPriceCommand{
		ProductID:    req.ProductID,
		PharmacyID:   req.PharmacyID,
		PharmacyName: req.PharmacyName,
		ProductName:  req.ProductName,
		Price:        req.Price,
		Source:       req.Source,
	}

	response, _ := mediator.Send[commands.RecordPriceCommand, responses.PriceHistoryItem](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}
