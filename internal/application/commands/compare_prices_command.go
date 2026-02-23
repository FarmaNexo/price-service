// internal/application/commands/compare_prices_command.go
package commands

// ComparePricesCommand comando para comparar precios de un producto
type ComparePricesCommand struct {
	ProductID string  `json:"product_id"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	RadiusKm  float64 `json:"radius_km,omitempty"`
}

func (c ComparePricesCommand) GetName() string {
	return "ComparePricesCommand"
}
