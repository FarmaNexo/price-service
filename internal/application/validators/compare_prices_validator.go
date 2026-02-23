// internal/application/validators/compare_prices_validator.go
package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/farmanexo/price-service/internal/application/commands"
	"github.com/farmanexo/price-service/internal/presentation/dto/responses"
	"github.com/farmanexo/price-service/pkg/mediator"
)

// ComparePricesValidator valida el comando de comparación de precios
type ComparePricesValidator struct{}

func NewComparePricesValidator() *ComparePricesValidator {
	return &ComparePricesValidator{}
}

func (v *ComparePricesValidator) Validate(ctx context.Context, cmd commands.ComparePricesCommand) error {
	var errors []string

	if strings.TrimSpace(cmd.ProductID) == "" {
		errors = append(errors, "El ID del producto es requerido")
	}

	if len(errors) > 0 {
		return fmt.Errorf("errores de validación: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Compile-time interface check
var _ mediator.Validator[commands.ComparePricesCommand, responses.PriceComparisonResponse] = (*ComparePricesValidator)(nil)
