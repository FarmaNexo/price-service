// internal/application/handlers/get_generic_vs_brand_handler.go
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

// GetGenericVsBrandHandler maneja la consulta de comparación genérico vs marca
type GetGenericVsBrandHandler struct {
	genericBrandRepo repositories.GenericBrandRepository
	cacheService     services.CacheService
	logger           *zap.Logger
}

func NewGetGenericVsBrandHandler(
	genericBrandRepo repositories.GenericBrandRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *GetGenericVsBrandHandler {
	return &GetGenericVsBrandHandler{
		genericBrandRepo: genericBrandRepo,
		cacheService:     cacheService,
		logger:           logger,
	}
}

func (h *GetGenericVsBrandHandler) Handle(ctx context.Context, query queries.GetGenericVsBrandQuery) (*common.ApiResponse[responses.GenericVsBrandResponse], error) {
	// Intentar caché
	cacheKey := fmt.Sprintf("cache:price:generic-vs-brand:%s", query.ProductID)
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var cachedResp responses.GenericVsBrandResponse
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			resp := common.OkResponse(cachedResp)
			resp.AddMessage(constants.CodeGenericVsBrandRetrieved, constants.GetDescription(constants.CodeGenericVsBrandRetrieved))
			return resp, nil
		}
	}

	comparisons, err := h.genericBrandRepo.FindByProductID(ctx, query.ProductID)
	if err != nil {
		h.logger.Error("Error obteniendo comparaciones genérico vs marca", zap.Error(err))
		return common.InternalServerErrorResponse[responses.GenericVsBrandResponse]("Error obteniendo comparaciones"), nil
	}

	items := make([]responses.GenericBrandItem, 0, len(comparisons))
	for _, c := range comparisons {
		items = append(items, responses.GenericBrandItem{
			ProductID:          c.ProductID,
			ProductName:        c.ProductName,
			IsGeneric:          c.IsGeneric,
			AvgPrice:           c.AvgPriceProduct,
			RelatedProductID:   c.RelatedProductID,
			RelatedProductName: c.RelatedProductName,
			RelatedIsGeneric:   c.RelatedIsGeneric,
			RelatedAvgPrice:    c.AvgPriceRelated,
			SavingsPercentage:  c.SavingsPercentage,
			ActiveIngredient:   c.ActiveIngredient,
		})
	}

	result := responses.GenericVsBrandResponse{
		ProductID:   query.ProductID,
		Comparisons: items,
		Total:       len(items),
	}

	// Guardar en caché (1 hora)
	if data, err := json.Marshal(result); err == nil {
		h.cacheService.Set(ctx, cacheKey, string(data), 1*time.Hour)
	}

	resp := common.OkResponse(result)
	resp.AddMessage(constants.CodeGenericVsBrandRetrieved, constants.GetDescription(constants.CodeGenericVsBrandRetrieved))
	return resp, nil
}

// Compile-time interface check
var _ mediator.RequestHandler[queries.GetGenericVsBrandQuery, responses.GenericVsBrandResponse] = (*GetGenericVsBrandHandler)(nil)
