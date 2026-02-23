// internal/application/queries/get_price_stats_query.go
package queries

// GetPriceStatsQuery consulta para obtener estadísticas de precios (admin)
type GetPriceStatsQuery struct {
	ProductID string `json:"product_id"`
}

func (q GetPriceStatsQuery) GetName() string {
	return "GetPriceStatsQuery"
}
