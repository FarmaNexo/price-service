// internal/domain/repositories/price_alert_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/price-service/internal/domain/entities"
)

// PriceAlertRepository interfaz para el repositorio de alertas de precio
type PriceAlertRepository interface {
	Create(ctx context.Context, alert *entities.PriceAlert) error
	FindByID(ctx context.Context, id string) (*entities.PriceAlert, error)
	FindByUserID(ctx context.Context, userID string) ([]entities.PriceAlert, error)
	FindActiveByProductID(ctx context.Context, productID string) ([]entities.PriceAlert, error)
	FindByUserAndProduct(ctx context.Context, userID, productID string) (*entities.PriceAlert, error)
	Update(ctx context.Context, alert *entities.PriceAlert) error
	Delete(ctx context.Context, id, userID string) error
	CountActiveByUser(ctx context.Context, userID string) (int64, error)
}
