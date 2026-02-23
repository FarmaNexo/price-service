// internal/presentation/dto/requests/price_requests.go
package requests

// ComparePricesRequest request para comparar precios
type ComparePricesRequest struct {
	ProductID string  `json:"product_id" example:"uuid"`
	Latitude  float64 `json:"latitude,omitempty" example:"-12.0464"`
	Longitude float64 `json:"longitude,omitempty" example:"-77.0428"`
	RadiusKm  float64 `json:"radius_km,omitempty" example:"5.0"`
}

// CreatePriceAlertRequest request para crear alerta de precio
type CreatePriceAlertRequest struct {
	ProductID   string  `json:"product_id" example:"uuid"`
	ProductName string  `json:"product_name" example:"Paracetamol 500mg"`
	TargetPrice float64 `json:"target_price" example:"5.50"`
}

// RecordPriceRequest request para registrar precio manualmente
type RecordPriceRequest struct {
	ProductID    string  `json:"product_id" example:"uuid"`
	PharmacyID   string  `json:"pharmacy_id" example:"uuid"`
	PharmacyName string  `json:"pharmacy_name" example:"Farmacia Central"`
	ProductName  string  `json:"product_name" example:"Paracetamol 500mg"`
	Price        float64 `json:"price" example:"5.50"`
	Source       string  `json:"source,omitempty" example:"manual"`
}
