// internal/application/validators/create_price_alert_validator.go
package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/farmanexo/price-service/internal/application/commands"
	"github.com/farmanexo/price-service/internal/presentation/dto/responses"
	"github.com/farmanexo/price-service/pkg/mediator"
)

// CreatePriceAlertValidator valida el comando de creación de alerta de precio
type CreatePriceAlertValidator struct{}

func NewCreatePriceAlertValidator() *CreatePriceAlertValidator {
	return &CreatePriceAlertValidator{}
}

func (v *CreatePriceAlertValidator) Validate(ctx context.Context, cmd commands.CreatePriceAlertCommand) error {
	var errors []string

	if strings.TrimSpace(cmd.ProductID) == "" {
		errors = append(errors, "El ID del producto es requerido")
	}
	if cmd.TargetPrice <= 0 {
		errors = append(errors, "El precio objetivo debe ser mayor a 0")
	}
	if cmd.UserID == "" {
		errors = append(errors, "El ID del usuario es requerido")
	}

	if len(errors) > 0 {
		return fmt.Errorf("errores de validación: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Compile-time interface check
var _ mediator.Validator[commands.CreatePriceAlertCommand, responses.PriceAlertResponse] = (*CreatePriceAlertValidator)(nil)
