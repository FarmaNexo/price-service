// internal/domain/services/pharmacy_client.go
package services

import "context"

// PharmacyInventoryItem representa un item del inventario de farmacia.
// DistanceKm se llena solo cuando GetProductPrices se invoca con geo (HU-014).
// HU-016: DistrictAvgPrice + IsOverpriced + OverpricePct viajan calculados
// desde pharmacy-service (la regla de "30% sobre el promedio del distrito"
// vive en su domain/services/pricing_rules.go). price-service solo propaga.
type PharmacyInventoryItem struct {
	PharmacyID       string   `json:"pharmacy_id"`
	PharmacyName     string   `json:"pharmacy_name"`
	ProductID        string   `json:"product_id"`
	Stock            int      `json:"stock"`
	Price            float64  `json:"price"`
	IsAvailable      bool     `json:"is_available"`
	DistanceKm       *float64 `json:"distance_km,omitempty"`
	DistrictAvgPrice *float64 `json:"district_avg_price,omitempty"`
	IsOverpriced     bool     `json:"is_overpriced"`
	OverpricePct     *float64 `json:"overprice_pct,omitempty"`
}

// PriceCompareGeo agrupa los parámetros opcionales de geolocalización para
// el endpoint de comparación de precios. Sin lat/lng → comportamiento legacy.
type PriceCompareGeo struct {
	Lat      float64
	Lng      float64
	RadiusKm float64
}

func (g PriceCompareGeo) IsActive() bool { return g.Lat != 0 || g.Lng != 0 }

// PharmacyClient interfaz para comunicación con Pharmacy Service
type PharmacyClient interface {
	GetProductPrices(ctx context.Context, productID string, geo PriceCompareGeo) ([]PharmacyInventoryItem, error)
	GetPharmacyInfo(ctx context.Context, pharmacyID string) (*PharmacyInfo, error)
}

// PharmacyInfo información básica de una farmacia
type PharmacyInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
