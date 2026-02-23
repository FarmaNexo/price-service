// internal/application/commands/create_price_alert_command.go
package commands

// CreatePriceAlertCommand comando para crear una alerta de precio
type CreatePriceAlertCommand struct {
	UserID      string  `json:"-"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	TargetPrice float64 `json:"target_price"`
}

func (c CreatePriceAlertCommand) GetName() string {
	return "CreatePriceAlertCommand"
}
