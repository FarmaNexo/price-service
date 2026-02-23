// internal/application/commands/record_price_command.go
package commands

// RecordPriceCommand comando para registrar un precio manualmente (admin)
type RecordPriceCommand struct {
	ProductID    string  `json:"product_id"`
	PharmacyID   string  `json:"pharmacy_id"`
	PharmacyName string  `json:"pharmacy_name"`
	ProductName  string  `json:"product_name"`
	Price        float64 `json:"price"`
	Source       string  `json:"source"`
}

func (c RecordPriceCommand) GetName() string {
	return "RecordPriceCommand"
}
