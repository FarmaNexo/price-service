// internal/infrastructure/clients/catalog_client_impl.go
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/farmanexo/price-service/internal/domain/services"
	"go.uber.org/zap"
)

// CatalogClientImpl implementa CatalogClient via HTTP.
// Timeout corto (3s) — fallar rápido y dejar que el handler decida si degrada gracefully.
type CatalogClientImpl struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewCatalogClient(baseURL string, logger *zap.Logger) *CatalogClientImpl {
	return &CatalogClientImpl{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		logger: logger,
	}
}

type catalogProductAPIResponse struct {
	Data *services.ProductInfo `json:"datos"`
}

// catalogSearchListResponse — el endpoint /products/search devuelve PaginatedResult,
// no una lista plana. Modelamos solo los campos que necesitamos.
type catalogSearchListResponse struct {
	Data struct {
		Products []services.ProductInfo `json:"products"`
	} `json:"datos"`
}

// catalogSearchRequest — body para POST /products/search.
type catalogSearchRequest struct {
	ActiveIngredient string `json:"active_ingredient,omitempty"`
	ExcludeID        string `json:"exclude_id,omitempty"`
	Page             int    `json:"page,omitempty"`
	Limit            int    `json:"limit,omitempty"`
}

func (c *CatalogClientImpl) GetProduct(ctx context.Context, productID string) (*services.ProductInfo, error) {
	url := fmt.Sprintf("%s/api/v1/products/%s", c.baseURL, productID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Error consultando Catalog Service",
			zap.String("url", url),
			zap.Error(err),
		)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog service retornó status %d", resp.StatusCode)
	}

	var apiResp catalogProductAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	return apiResp.Data, nil
}

// GetProductsByActiveIngredient busca productos con la misma DCI vía POST /products/search
// con body JSON. Si el servicio no responde o falla, retorna lista vacía (graceful degradation).
func (c *CatalogClientImpl) GetProductsByActiveIngredient(
	ctx context.Context,
	activeIngredient, excludeID string,
	limit int,
) ([]services.ProductInfo, error) {
	if activeIngredient == "" {
		return []services.ProductInfo{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	url := fmt.Sprintf("%s/api/v1/products/search", c.baseURL)
	body := catalogSearchRequest{
		ActiveIngredient: activeIngredient,
		ExcludeID:        excludeID,
		Page:             1,
		Limit:            limit,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return []services.ProductInfo{}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return []services.ProductInfo{}, nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Error consultando Catalog Service por ingrediente activo",
			zap.String("url", url),
			zap.String("active_ingredient", activeIngredient),
			zap.Error(err),
		)
		return []services.ProductInfo{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Catalog Service retornó status no-OK al buscar por DCI",
			zap.Int("status", resp.StatusCode),
			zap.String("active_ingredient", activeIngredient),
		)
		return []services.ProductInfo{}, nil
	}

	var apiResp catalogSearchListResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		c.logger.Warn("Error decodificando respuesta de búsqueda",
			zap.Error(err),
		)
		return []services.ProductInfo{}, nil
	}

	return apiResp.Data.Products, nil
}

// Compile-time interface check
var _ services.CatalogClient = (*CatalogClientImpl)(nil)
