// internal/infrastructure/clients/catalog_client_impl.go
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/farmanexo/price-service/internal/domain/services"
	"go.uber.org/zap"
)

// CatalogClientImpl implementa CatalogClient via HTTP
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

type catalogProductListAPIResponse struct {
	Data []services.ProductInfo `json:"datos"`
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

func (c *CatalogClientImpl) GetProductsByActiveIngredient(ctx context.Context, activeIngredient string) ([]services.ProductInfo, error) {
	url := fmt.Sprintf("%s/api/v1/products/search", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	q := req.URL.Query()
	q.Set("active_ingredient", activeIngredient)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Error consultando Catalog Service por ingrediente activo",
			zap.String("url", url),
			zap.Error(err),
		)
		return []services.ProductInfo{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []services.ProductInfo{}, nil
	}

	var apiResp catalogProductListAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return []services.ProductInfo{}, nil
	}

	return apiResp.Data, nil
}

// Compile-time interface check
var _ services.CatalogClient = (*CatalogClientImpl)(nil)
