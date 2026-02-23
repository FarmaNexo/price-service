// internal/application/queries/get_price_history_query.go
package queries

// GetPriceHistoryQuery consulta para obtener el historial de precios
type GetPriceHistoryQuery struct {
	ProductID  string `json:"product_id"`
	PharmacyID string `json:"pharmacy_id,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

func (q GetPriceHistoryQuery) GetName() string {
	return "GetPriceHistoryQuery"
}
