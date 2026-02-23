// internal/domain/repositories/generic_brand_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/price-service/internal/domain/entities"
)

// GenericBrandRepository interfaz para el repositorio de comparaciones genérico vs marca
type GenericBrandRepository interface {
	Create(ctx context.Context, comparison *entities.GenericBrandComparison) error
	FindByProductID(ctx context.Context, productID string) ([]entities.GenericBrandComparison, error)
	FindByActiveIngredient(ctx context.Context, activeIngredient string) ([]entities.GenericBrandComparison, error)
	Update(ctx context.Context, comparison *entities.GenericBrandComparison) error
	Upsert(ctx context.Context, comparison *entities.GenericBrandComparison) error
}
