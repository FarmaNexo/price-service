// internal/application/queries/list_price_alerts_query.go
package queries

// ListPriceAlertsQuery consulta para listar alertas de precio del usuario
type ListPriceAlertsQuery struct {
	UserID string `json:"-"`
}

func (q ListPriceAlertsQuery) GetName() string {
	return "ListPriceAlertsQuery"
}
