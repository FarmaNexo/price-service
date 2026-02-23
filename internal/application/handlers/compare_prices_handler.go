// internal/application/handlers/compare_prices_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/farmanexo/price-service/internal/application/commands"
	"github.com/farmanexo/price-service/internal/domain/repositories"
	"github.com/farmanexo/price-service/internal/domain/services"
	"github.com/farmanexo/price-service/internal/presentation/dto/responses"
	"github.com/farmanexo/price-service/internal/shared/common"
	"github.com/farmanexo/price-service/internal/shared/constants"
	"github.com/farmanexo/price-service/pkg/mediator"
	"go.uber.org/zap"
)

// ComparePricesHandler maneja la comparación de precios
type ComparePricesHandler struct {
	priceHistoryRepo repositories.PriceHistoryRepository
	pharmacyClient   services.PharmacyClient
	cacheService     services.CacheService
	logger           *zap.Logger
}

func NewComparePricesHandler(
	priceHistoryRepo repositories.PriceHistoryRepository,
	pharmacyClient services.PharmacyClient,
	cacheService services.CacheService,
	logger *zap.Logger,
) *ComparePricesHandler {
	return &ComparePricesHandler{
		priceHistoryRepo: priceHistoryRepo,
		pharmacyClient:   pharmacyClient,
		cacheService:     cacheService,
		logger:           logger,
	}
}

func (h *ComparePricesHandler) Handle(ctx context.Context, cmd commands.ComparePricesCommand) (*common.ApiResponse[responses.PriceComparisonResponse], error) {
	// Intentar obtener de caché
	cacheKey := fmt.Sprintf("cache:price:compare:%s", cmd.ProductID)
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var cachedResp responses.PriceComparisonResponse
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			h.logger.Debug("Comparación de precios obtenida de caché", zap.String("product_id", cmd.ProductID))
			resp := common.OkResponse(cachedResp)
			resp.AddMessage(constants.CodePriceCompared, constants.GetDescription(constants.CodePriceCompared))
			return resp, nil
		}
	}

	// Obtener precios actuales del Pharmacy Service
	inventoryItems, err := h.pharmacyClient.GetProductPrices(ctx, cmd.ProductID)
	if err != nil {
		h.logger.Warn("Error obteniendo precios de farmacias", zap.Error(err))
	}

	// Construir respuesta de comparación
	pharmacyPrices := make([]responses.PharmacyPriceItem, 0, len(inventoryItems))
	var minPrice, maxPrice float64
	var totalPrice float64

	for i, item := range inventoryItems {
		if !item.IsAvailable || item.Stock <= 0 {
			continue
		}

		pharmacyPrices = append(pharmacyPrices, responses.PharmacyPriceItem{
			PharmacyID:   item.PharmacyID,
			PharmacyName: item.PharmacyName,
			Price:        item.Price,
			Stock:        item.Stock,
			IsAvailable:  item.IsAvailable,
		})

		if i == 0 || item.Price < minPrice {
			minPrice = item.Price
		}
		if item.Price > maxPrice {
			maxPrice = item.Price
		}
		totalPrice += item.Price
	}

	avgPrice := float64(0)
	if len(pharmacyPrices) > 0 {
		avgPrice = totalPrice / float64(len(pharmacyPrices))
	}

	comparison := responses.PriceComparisonResponse{
		ProductID:       cmd.ProductID,
		PharmacyPrices:  pharmacyPrices,
		MinPrice:        minPrice,
		MaxPrice:        maxPrice,
		AvgPrice:        avgPrice,
		TotalPharmacies: len(pharmacyPrices),
	}

	// Guardar en caché (15 minutos)
	if data, err := json.Marshal(comparison); err == nil {
		h.cacheService.Set(ctx, cacheKey, string(data), 15*time.Minute)
	}

	resp := common.OkResponse(comparison)
	resp.AddMessage(constants.CodePriceCompared, constants.GetDescription(constants.CodePriceCompared))
	return resp, nil
}

// Compile-time interface check
var _ mediator.RequestHandler[commands.ComparePricesCommand, responses.PriceComparisonResponse] = (*ComparePricesHandler)(nil)
