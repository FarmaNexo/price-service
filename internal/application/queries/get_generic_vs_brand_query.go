// internal/application/queries/get_generic_vs_brand_query.go
package queries

// GetGenericVsBrandQuery consulta para obtener comparación genérico vs marca
type GetGenericVsBrandQuery struct {
	ProductID string `json:"product_id"`
}

func (q GetGenericVsBrandQuery) GetName() string {
	return "GetGenericVsBrandQuery"
}
