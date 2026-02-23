// internal/application/handlers/get_price_stats_handler.go
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

// GetPriceStatsHandler maneja la consulta de estadísticas de precios (admin)
type GetPriceStatsHandler struct {
	priceHistoryRepo repositories.PriceHistoryRepository
	cacheService     services.CacheService
	logger           *zap.Logger
}

func NewGetPriceStatsHandler(
	priceHistoryRepo repositories.PriceHistoryRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *GetPriceStatsHandler {
	return &GetPriceStatsHandler{
		priceHistoryRepo: priceHistoryRepo,
		cacheService:     cacheService,
		logger:           logger,
	}
}

func (h *GetPriceStatsHandler) Handle(ctx context.Context, query queries.GetPriceStatsQuery) (*common.ApiResponse[responses.PriceStatsResponse], error) {
	// Intentar caché
	cacheKey := fmt.Sprintf("cache:price:stats:%s", query.ProductID)
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var cachedResp responses.PriceStatsResponse
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			resp := common.OkResponse(cachedResp)
			resp.AddMessage(constants.CodePriceStatsRetrieved, constants.GetDescription(constants.CodePriceStatsRetrieved))
			return resp, nil
		}
	}

	avgPrice, err := h.priceHistoryRepo.GetAvgPrice(ctx, query.ProductID)
	if err != nil {
		h.logger.Error("Error obteniendo precio promedio", zap.Error(err))
	}

	var minPriceVal, maxPriceVal float64
	var minPharmacy, maxPharmacy string

	minRecord, err := h.priceHistoryRepo.GetMinPrice(ctx, query.ProductID)
	if err == nil && minRecord != nil {
		minPriceVal = minRecord.Price
		minPharmacy = minRecord.PharmacyName
	}

	maxRecord, err := h.priceHistoryRepo.GetMaxPrice(ctx, query.ProductID)
	if err == nil && maxRecord != nil {
		maxPriceVal = maxRecord.Price
		maxPharmacy = maxRecord.PharmacyName
	}

	latestPrices, err := h.priceHistoryRepo.FindLatestByProduct(ctx, query.ProductID)
	if err != nil {
		h.logger.Error("Error obteniendo precios más recientes", zap.Error(err))
	}

	result := responses.PriceStatsResponse{
		ProductID:        query.ProductID,
		AvgPrice:         avgPrice,
		MinPrice:         minPriceVal,
		MinPricePharmacy: minPharmacy,
		MaxPrice:         maxPriceVal,
		MaxPricePharmacy: maxPharmacy,
		TotalRecords:     len(latestPrices),
	}

	// Guardar en caché (30 minutos)
	if data, err := json.Marshal(result); err == nil {
		h.cacheService.Set(ctx, cacheKey, string(data), 30*time.Minute)
	}

	resp := common.OkResponse(result)
	resp.AddMessage(constants.CodePriceStatsRetrieved, constants.GetDescription(constants.CodePriceStatsRetrieved))
	return resp, nil
}

// Compile-time interface check
var _ mediator.RequestHandler[queries.GetPriceStatsQuery, responses.PriceStatsResponse] = (*GetPriceStatsHandler)(nil)
