// internal/application/handlers/get_product_alternatives_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/farmanexo/price-service/internal/application/queries"
	"github.com/farmanexo/price-service/internal/domain/services"
	"github.com/farmanexo/price-service/internal/presentation/dto/responses"
	"github.com/farmanexo/price-service/internal/shared/common"
	"github.com/farmanexo/price-service/internal/shared/constants"
	"github.com/farmanexo/price-service/pkg/mediator"
	"go.uber.org/zap"
)

// GetProductAlternativesHandler — HU-015. Computa alternativas terapéuticas en tiempo real:
//   1. Carga el producto base desde Catalog (para obtener su DCI).
//   2. Pide a Catalog los productos con la misma DCI (excluyendo el base).
//   3. Para cada candidato, consulta a Pharmacy las farmacias que lo venden y calcula el precio promedio.
//   4. Calcula % de ahorro vs el precio promedio del producto base.
//   5. Retorna ordenados por mayor ahorro primero, hasta el límite solicitado.
//
// Diseño:
//   - Cache Redis 30 min (precios cambian con eventos INVENTORY_UPDATED, no requieren TTL bajo).
//   - Graceful degradation: si Catalog o Pharmacy fallan, retorna respuesta vacía con 200, no 5xx.
//   - Solo alternativas con al menos 1 farmacia con stock se incluyen (evita sugerir productos no comprables).
type GetProductAlternativesHandler struct {
	catalogClient  services.CatalogClient
	pharmacyClient services.PharmacyClient
	cacheService   services.CacheService
	logger         *zap.Logger
}

func NewGetProductAlternativesHandler(
	catalogClient services.CatalogClient,
	pharmacyClient services.PharmacyClient,
	cacheService services.CacheService,
	logger *zap.Logger,
) *GetProductAlternativesHandler {
	return &GetProductAlternativesHandler{
		catalogClient:  catalogClient,
		pharmacyClient: pharmacyClient,
		cacheService:   cacheService,
		logger:         logger,
	}
}

const (
	alternativesDefaultLimit = 10
	alternativesMaxLimit     = 25
	alternativesCacheTTL     = 30 * time.Minute
	alternativesCandidatePoolSize = 20 // Cuántos candidatos pedir a catalog antes de filtrar por precio.
)

func (h *GetProductAlternativesHandler) Handle(
	ctx context.Context,
	query queries.GetProductAlternativesQuery,
) (*common.ApiResponse[responses.ProductAlternativesResponse], error) {

	limit := query.Limit
	if limit <= 0 {
		limit = alternativesDefaultLimit
	}
	if limit > alternativesMaxLimit {
		limit = alternativesMaxLimit
	}

	cacheKey := fmt.Sprintf("cache:price:alternatives:%s:limit-%d", query.ProductID, limit)
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		cachedResp := responses.ProductAlternativesResponse{}
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			resp := common.OkResponse(cachedResp)
			resp.AddMessage(constants.CodeAlternativesRetrieved, constants.GetDescription(constants.CodeAlternativesRetrieved))
			return resp, nil
		}
	}

	// 1. Producto base
	baseProduct, err := h.catalogClient.GetProduct(ctx, query.ProductID)
	if err != nil || baseProduct == nil {
		h.logger.Warn("No se pudo cargar el producto base desde Catalog",
			zap.String("product_id", query.ProductID),
			zap.Error(err),
		)
		return h.emptyResponse(query.ProductID, "", "", 0), nil
	}
	if baseProduct.ActiveIngredient == "" {
		h.logger.Debug("Producto sin ingrediente activo, no hay alternativas posibles",
			zap.String("product_id", query.ProductID),
		)
		return h.emptyResponse(baseProduct.ID, baseProduct.Name, "", 0), nil
	}

	// 2. Precio promedio del producto base
	baseAvgPrice, _ := h.computeAvgPrice(ctx, baseProduct.ID)

	// 3. Candidatos con misma DCI
	candidates, err := h.catalogClient.GetProductsByActiveIngredient(
		ctx, baseProduct.ActiveIngredient, baseProduct.ID, alternativesCandidatePoolSize,
	)
	if err != nil {
		h.logger.Warn("Error consultando alternativas en catalog (graceful empty)",
			zap.String("product_id", query.ProductID),
			zap.Error(err),
		)
		return h.emptyResponse(baseProduct.ID, baseProduct.Name, baseProduct.ActiveIngredient, baseAvgPrice), nil
	}

	// 4. Para cada candidato, calcular avg price + savings
	items := make([]responses.ProductAlternativeItem, 0, len(candidates))
	for _, c := range candidates {
		avgPrice, pharmaciesCount := h.computeAvgPrice(ctx, c.ID)
		if pharmaciesCount == 0 || avgPrice <= 0 {
			// No hay farmacias con stock disponible — no la sugerimos.
			continue
		}

		savings := 0.0
		if baseAvgPrice > 0 {
			savings = ((baseAvgPrice - avgPrice) / baseAvgPrice) * 100
		}

		items = append(items, responses.ProductAlternativeItem{
			ProductID:         c.ID,
			ProductName:       c.Name,
			ProductSlug:       c.Slug,
			IsGeneric:         c.IsGeneric,
			Manufacturer:      c.Manufacturer,
			Presentation:      c.Presentation,
			AvgPrice:          round2(avgPrice),
			PharmaciesCount:   pharmaciesCount,
			SavingsPercentage: round2(savings),
		})
	}

	// 5. Orden por ahorro DESC + cap al limit solicitado
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].SavingsPercentage > items[j].SavingsPercentage
	})
	if len(items) > limit {
		items = items[:limit]
	}

	result := responses.ProductAlternativesResponse{
		BaseProductID:    baseProduct.ID,
		BaseProductName:  baseProduct.Name,
		BaseAvgPrice:     round2(baseAvgPrice),
		ActiveIngredient: baseProduct.ActiveIngredient,
		Alternatives:     items,
		Total:            len(items),
	}

	if data, err := json.Marshal(result); err == nil {
		_ = h.cacheService.Set(ctx, cacheKey, string(data), alternativesCacheTTL)
	}

	resp := common.OkResponse(result)
	resp.AddMessage(constants.CodeAlternativesRetrieved, constants.GetDescription(constants.CodeAlternativesRetrieved))
	return resp, nil
}

// computeAvgPrice consulta Pharmacy y devuelve (avg_price, pharmacies_count) considerando
// solo items con is_available=true y price > 0. Si Pharmacy falla o no hay datos, retorna (0, 0).
func (h *GetProductAlternativesHandler) computeAvgPrice(ctx context.Context, productID string) (float64, int) {
	// HU-015 calcula avg_price agnóstico de la ubicación del usuario;
	// pasamos geo vacío explícitamente para no filtrar por radio.
	items, err := h.pharmacyClient.GetProductPrices(ctx, productID, services.PriceCompareGeo{})
	if err != nil {
		return 0, 0
	}
	sum := 0.0
	count := 0
	for _, it := range items {
		if !it.IsAvailable || it.Price <= 0 {
			continue
		}
		sum += it.Price
		count++
	}
	if count == 0 {
		return 0, 0
	}
	return sum / float64(count), count
}

func (h *GetProductAlternativesHandler) emptyResponse(
	productID, productName, ingredient string, baseAvg float64,
) *common.ApiResponse[responses.ProductAlternativesResponse] {
	result := responses.ProductAlternativesResponse{
		BaseProductID:    productID,
		BaseProductName:  productName,
		BaseAvgPrice:     round2(baseAvg),
		ActiveIngredient: ingredient,
		Alternatives:     []responses.ProductAlternativeItem{},
		Total:            0,
	}
	resp := common.OkResponse(result)
	resp.AddMessage(constants.CodeAlternativesRetrieved, constants.GetDescription(constants.CodeAlternativesRetrieved))
	return resp
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

// Compile-time interface check
var _ mediator.RequestHandler[queries.GetProductAlternativesQuery, responses.ProductAlternativesResponse] = (*GetProductAlternativesHandler)(nil)
