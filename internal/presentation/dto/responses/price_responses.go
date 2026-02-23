// internal/presentation/dto/responses/price_responses.go
package responses

import (
	"github.com/farmanexo/price-service/internal/domain/entities"
)

// ========================================
// PRICE COMPARISON
// ========================================

// PriceComparisonResponse respuesta de comparación de precios
type PriceComparisonResponse struct {
	ProductID       string              `json:"product_id"`
	PharmacyPrices  []PharmacyPriceItem `json:"pharmacy_prices"`
	MinPrice        float64             `json:"min_price"`
	MaxPrice        float64             `json:"max_price"`
	AvgPrice        float64             `json:"avg_price"`
	TotalPharmacies int                 `json:"total_pharmacies"`
}

// PharmacyPriceItem precio por farmacia
type PharmacyPriceItem struct {
	PharmacyID   string  `json:"pharmacy_id"`
	PharmacyName string  `json:"pharmacy_name"`
	Price        float64 `json:"price"`
	Stock        int     `json:"stock"`
	IsAvailable  bool    `json:"is_available"`
}

// ========================================
// PRICE HISTORY
// ========================================

// PriceHistoryListResponse respuesta de historial de precios
type PriceHistoryListResponse struct {
	ProductID string             `json:"product_id"`
	History   []PriceHistoryItem `json:"history"`
	Total     int                `json:"total"`
}

// PriceHistoryItem item del historial de precios
type PriceHistoryItem struct {
	ID            string   `json:"id"`
	ProductID     string   `json:"product_id"`
	PharmacyID    string   `json:"pharmacy_id"`
	PharmacyName  string   `json:"pharmacy_name"`
	ProductName   string   `json:"product_name"`
	Price         float64  `json:"price"`
	PreviousPrice *float64 `json:"previous_price,omitempty"`
	Currency      string   `json:"currency"`
	Source        string   `json:"source"`
	RecordedAt    string   `json:"recorded_at"`
}

// ToPriceHistoryItem convierte entidad a response
func ToPriceHistoryItem(h entities.PriceHistory) PriceHistoryItem {
	return PriceHistoryItem{
		ID:            h.ID,
		ProductID:     h.ProductID,
		PharmacyID:    h.PharmacyID,
		PharmacyName:  h.PharmacyName,
		ProductName:   h.ProductName,
		Price:         h.Price,
		PreviousPrice: h.PreviousPrice,
		Currency:      h.Currency,
		Source:        h.Source,
		RecordedAt:    h.RecordedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ========================================
// PRICE ALERTS
// ========================================

// PriceAlertListResponse respuesta de lista de alertas
type PriceAlertListResponse struct {
	Alerts []PriceAlertResponse `json:"alerts"`
	Total  int                  `json:"total"`
}

// PriceAlertResponse respuesta de alerta de precio
type PriceAlertResponse struct {
	ID                    string   `json:"id"`
	UserID                string   `json:"user_id"`
	ProductID             string   `json:"product_id"`
	ProductName           string   `json:"product_name"`
	TargetPrice           float64  `json:"target_price"`
	CurrentPrice          *float64 `json:"current_price,omitempty"`
	IsActive              bool     `json:"is_active"`
	IsTriggered           bool     `json:"is_triggered"`
	TriggeredAt           *string  `json:"triggered_at,omitempty"`
	TriggeredPharmacyID   *string  `json:"triggered_pharmacy_id,omitempty"`
	TriggeredPharmacyName *string  `json:"triggered_pharmacy_name,omitempty"`
	TriggeredPrice        *float64 `json:"triggered_price,omitempty"`
	CreatedAt             string   `json:"created_at"`
}

// ToPriceAlertResponse convierte entidad a response
func ToPriceAlertResponse(a entities.PriceAlert) PriceAlertResponse {
	resp := PriceAlertResponse{
		ID:                    a.ID,
		UserID:                a.UserID,
		ProductID:             a.ProductID,
		ProductName:           a.ProductName,
		TargetPrice:           a.TargetPrice,
		CurrentPrice:          a.CurrentPrice,
		IsActive:              a.IsActive,
		IsTriggered:           a.IsTriggered,
		TriggeredPharmacyID:   a.TriggeredPharmacyID,
		TriggeredPharmacyName: a.TriggeredPharmacyName,
		TriggeredPrice:        a.TriggeredPrice,
		CreatedAt:             a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if a.TriggeredAt != nil {
		t := a.TriggeredAt.Format("2006-01-02T15:04:05Z")
		resp.TriggeredAt = &t
	}
	return resp
}

// DeleteAlertResponse respuesta de eliminación de alerta
type DeleteAlertResponse struct {
	AlertID string `json:"alert_id"`
	Deleted bool   `json:"deleted"`
}

// ========================================
// GENERIC VS BRAND
// ========================================

// GenericVsBrandResponse respuesta de comparación genérico vs marca
type GenericVsBrandResponse struct {
	ProductID   string             `json:"product_id"`
	Comparisons []GenericBrandItem `json:"comparisons"`
	Total       int                `json:"total"`
}

// GenericBrandItem item de comparación genérico vs marca
type GenericBrandItem struct {
	ProductID          string  `json:"product_id"`
	ProductName        string  `json:"product_name"`
	IsGeneric          bool    `json:"is_generic"`
	AvgPrice           float64 `json:"avg_price"`
	RelatedProductID   string  `json:"related_product_id"`
	RelatedProductName string  `json:"related_product_name"`
	RelatedIsGeneric   bool    `json:"related_is_generic"`
	RelatedAvgPrice    float64 `json:"related_avg_price"`
	SavingsPercentage  float64 `json:"savings_percentage"`
	ActiveIngredient   string  `json:"active_ingredient"`
}

// ========================================
// PRICE STATS
// ========================================

// PriceStatsResponse respuesta de estadísticas de precios
type PriceStatsResponse struct {
	ProductID        string  `json:"product_id"`
	AvgPrice         float64 `json:"avg_price"`
	MinPrice         float64 `json:"min_price"`
	MinPricePharmacy string  `json:"min_price_pharmacy"`
	MaxPrice         float64 `json:"max_price"`
	MaxPricePharmacy string  `json:"max_price_pharmacy"`
	TotalRecords     int     `json:"total_records"`
}
