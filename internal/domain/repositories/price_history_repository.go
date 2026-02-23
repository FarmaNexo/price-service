// internal/domain/repositories/price_history_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/price-service/internal/domain/entities"
)

// PriceHistoryRepository interfaz para el repositorio de historial de precios
type PriceHistoryRepository interface {
	Create(ctx context.Context, history *entities.PriceHistory) error
	FindByProductID(ctx context.Context, productID string, limit int) ([]entities.PriceHistory, error)
	FindByProductAndPharmacy(ctx context.Context, productID, pharmacyID string, limit int) ([]entities.PriceHistory, error)
	FindLatestByProduct(ctx context.Context, productID string) ([]entities.PriceHistory, error)
	FindLatestPrice(ctx context.Context, productID, pharmacyID string) (*entities.PriceHistory, error)
	GetAvgPrice(ctx context.Context, productID string) (float64, error)
	GetMinPrice(ctx context.Context, productID string) (*entities.PriceHistory, error)
	GetMaxPrice(ctx context.Context, productID string) (*entities.PriceHistory, error)
}
