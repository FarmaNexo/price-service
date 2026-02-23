// internal/application/handlers/get_price_history_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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

// GetPriceHistoryHandler maneja la consulta de historial de precios
type GetPriceHistoryHandler struct {
	priceHistoryRepo repositories.PriceHistoryRepository
	cacheService     services.CacheService
	logger           *zap.Logger
}

func NewGetPriceHistoryHandler(
	priceHistoryRepo repositories.PriceHistoryRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *GetPriceHistoryHandler {
	return &GetPriceHistoryHandler{
		priceHistoryRepo: priceHistoryRepo,
		cacheService:     cacheService,
		logger:           logger,
	}
}

func (h *GetPriceHistoryHandler) Handle(ctx context.Context, query queries.GetPriceHistoryQuery) (*common.ApiResponse[responses.PriceHistoryListResponse], error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}

	// Intentar caché
	cacheKey := fmt.Sprintf("cache:price:history:%s:%s:%d", query.ProductID, query.PharmacyID, limit)
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var cachedResp responses.PriceHistoryListResponse
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			h.logger.Debug("Historial de precios obtenido de caché", zap.String("product_id", query.ProductID))
			resp := common.OkResponse(cachedResp)
			resp.AddMessage(constants.CodePriceHistoryRetrieved, constants.GetDescription(constants.CodePriceHistoryRetrieved))
			return resp, nil
		}
	}

	var historyItems []responses.PriceHistoryItem

	if query.PharmacyID != "" {
		records, err := h.priceHistoryRepo.FindByProductAndPharmacy(ctx, query.ProductID, query.PharmacyID, limit)
		if err != nil {
			h.logger.Error("Error obteniendo historial por producto y farmacia", zap.Error(err))
			return common.InternalServerErrorResponse[responses.PriceHistoryListResponse]("Error obteniendo historial de precios"), nil
		}
		for _, r := range records {
			historyItems = append(historyItems, responses.ToPriceHistoryItem(r))
		}
	} else {
		records, err := h.priceHistoryRepo.FindByProductID(ctx, query.ProductID, limit)
		if err != nil {
			h.logger.Error("Error obteniendo historial por producto", zap.Error(err))
			return common.InternalServerErrorResponse[responses.PriceHistoryListResponse]("Error obteniendo historial de precios"), nil
		}
		for _, r := range records {
			historyItems = append(historyItems, responses.ToPriceHistoryItem(r))
		}
	}

	if historyItems == nil {
		historyItems = []responses.PriceHistoryItem{}
	}

	result := responses.PriceHistoryListResponse{
		ProductID: query.ProductID,
		History:   historyItems,
		Total:     len(historyItems),
	}

	// Guardar en caché (30 minutos)
	if data, err := json.Marshal(result); err == nil {
		h.cacheService.Set(ctx, cacheKey, string(data), 30*time.Minute)
	}

	resp := common.OkResponse(result)
	resp.AddMessage(constants.CodePriceHistoryRetrieved, constants.GetDescription(constants.CodePriceHistoryRetrieved))
	return resp, nil
}

// Compile-time interface check
var _ mediator.RequestHandler[queries.GetPriceHistoryQuery, responses.PriceHistoryListResponse] = (*GetPriceHistoryHandler)(nil)
