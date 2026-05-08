// internal/application/queries/get_product_alternatives_query.go
package queries

// GetProductAlternativesQuery — busca alternativas terapéuticas (mismo principio activo / DCI)
// para un producto, calculando precio promedio por alternativa y % de ahorro vs el producto base.
//
// Casos de uso:
//   - HU-015 — usuario busca medicamento de marca, plataforma sugiere genéricos equivalentes.
//   - Educación de paciente sobre opciones más económicas con bioequivalencia.
//
// Reglas:
//   - El producto base nunca está incluido en los resultados.
//   - Solo se consideran alternativas con precio disponible (al menos una farmacia con stock).
//   - Resultados ordenados por % de ahorro descendente (mejor opción primero).
//   - Limit por defecto: 10 alternativas.
type GetProductAlternativesQuery struct {
	ProductID string
	Limit     int
}

func (q GetProductAlternativesQuery) GetName() string {
	return "GetProductAlternativesQuery"
}
