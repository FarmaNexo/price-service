// internal/application/handlers/list_price_alerts_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/farmanexo/price-service/internal/application/queries"
	"github.com/farmanexo/price-service/internal/domain/repositories"
	"github.com/farmanexo/price-service/internal/domain/services"
	"github.com/farmanexo/price-service/internal/presentation/dto/responses"
	"github.com/farmanexo/price-service/internal/shared/common"
	"github.com/farmanexo/price-service/internal/shared/constants"
	"github.com/farmanexo/price-service/pkg/mediator"
	"go.uber.org/zap"
)

// ListPriceAlertsHandler maneja la consulta de alertas de precio del usuario
type ListPriceAlertsHandler struct {
	alertRepo    repositories.PriceAlertRepository
	cacheService services.CacheService
	logger       *zap.Logger
}

func NewListPriceAlertsHandler(
	alertRepo repositories.PriceAlertRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *ListPriceAlertsHandler {
	return &ListPriceAlertsHandler{
		alertRepo:    alertRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (h *ListPriceAlertsHandler) Handle(ctx context.Context, query queries.ListPriceAlertsQuery) (*common.ApiResponse[responses.PriceAlertListResponse], error) {
	// Intentar caché
	cacheKey := "cache:price:alerts:" + query.UserID
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var cachedResp responses.PriceAlertListResponse
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			resp := common.OkResponse(cachedResp)
			resp.AddMessage(constants.CodeAlertsListed, constants.GetDescription(constants.CodeAlertsListed))
			return resp, nil
		}
	}

	alerts, err := h.alertRepo.FindByUserID(ctx, query.UserID)
	if err != nil {
		h.logger.Error("Error obteniendo alertas del usuario", zap.Error(err))
		return common.InternalServerErrorResponse[responses.PriceAlertListResponse]("Error obteniendo alertas"), nil
	}

	alertResponses := make([]responses.PriceAlertResponse, 0, len(alerts))
	for _, alert := range alerts {
		alertResponses = append(alertResponses, responses.ToPriceAlertResponse(alert))
	}

	result := responses.PriceAlertListResponse{
		Alerts: alertResponses,
		Total:  len(alertResponses),
	}

	// Guardar en caché (10 minutos)
	if data, err := json.Marshal(result); err == nil {
		h.cacheService.Set(ctx, cacheKey, string(data), 10*time.Minute)
	}

	resp := common.OkResponse(result)
	resp.AddMessage(constants.CodeAlertsListed, constants.GetDescription(constants.CodeAlertsListed))
	return resp, nil
}

// Compile-time interface check
var _ mediator.RequestHandler[queries.ListPriceAlertsQuery, responses.PriceAlertListResponse] = (*ListPriceAlertsHandler)(nil)
