// internal/application/commands/delete_price_alert_command.go
package commands

// DeletePriceAlertCommand comando para eliminar una alerta de precio
type DeletePriceAlertCommand struct {
	AlertID string `json:"alert_id"`
	UserID  string `json:"-"`
}

func (c DeletePriceAlertCommand) GetName() string {
	return "DeletePriceAlertCommand"
}
