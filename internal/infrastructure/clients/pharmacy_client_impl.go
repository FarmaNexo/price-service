// internal/infrastructure/clients/pharmacy_client_impl.go
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"strconv"
	"time"

	"github.com/farmanexo/price-service/internal/domain/services"
	"go.uber.org/zap"
)

// PharmacyClientImpl implementa PharmacyClient via HTTP
type PharmacyClientImpl struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewPharmacyClient(baseURL string, logger *zap.Logger) *PharmacyClientImpl {
	return &PharmacyClientImpl{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		logger: logger,
	}
}

// Pharmacy Service responde con ApiResponse[InventoryListResponse]:
//   { "datos": { "items": [...], "total": N }, "meta": {...} }
type pharmacyInventoryAPIResponse struct {
	Data struct {
		Items []services.PharmacyInventoryItem `json:"items"`
		Total int                              `json:"total"`
	} `json:"datos"`
}

type pharmacyInfoAPIResponse struct {
	Data *services.PharmacyInfo `json:"datos"`
}

func (c *PharmacyClientImpl) GetProductPrices(ctx context.Context, productID string, geo services.PriceCompareGeo) ([]services.PharmacyInventoryItem, error) {
	// Endpoint correcto en pharmacy-service: /api/v1/pharmacies/inventory/product/{productId}
	// (ver pharmacy-service routes — devuelve InventoryListResponse con farmacia + slug + precio).
	// HU-014: si la query trae geo, propaga lat/lng/radius_km como query params.
	url := fmt.Sprintf("%s/api/v1/pharmacies/inventory/product/%s", c.baseURL, productID)
	if geo.IsActive() {
		q := neturl.Values{}
		q.Set("lat", strconv.FormatFloat(geo.Lat, 'f', -1, 64))
		q.Set("lng", strconv.FormatFloat(geo.Lng, 'f', -1, 64))
		if geo.RadiusKm > 0 {
			q.Set("radius_km", strconv.FormatFloat(geo.RadiusKm, 'f', -1, 64))
		}
		url = url + "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Error consultando Pharmacy Service",
			zap.String("url", url),
			zap.Error(err),
		)
		return []services.PharmacyInventoryItem{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Pharmacy Service retornó error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("url", url),
		)
		return []services.PharmacyInventoryItem{}, nil
	}

	var apiResp pharmacyInventoryAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		c.logger.Error("Error decodificando respuesta de Pharmacy Service", zap.Error(err))
		return []services.PharmacyInventoryItem{}, nil
	}

	return apiResp.Data.Items, nil
}

func (c *PharmacyClientImpl) GetPharmacyInfo(ctx context.Context, pharmacyID string) (*services.PharmacyInfo, error) {
	url := fmt.Sprintf("%s/api/v1/pharmacies/%s", c.baseURL, pharmacyID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Error consultando Pharmacy Service info",
			zap.String("url", url),
			zap.Error(err),
		)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pharmacy service retornó status %d", resp.StatusCode)
	}

	var apiResp pharmacyInfoAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	return apiResp.Data, nil
}

// Compile-time interface check
var _ services.PharmacyClient = (*PharmacyClientImpl)(nil)
