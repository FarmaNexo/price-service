// internal/infrastructure/persistence/postgres/price_history_repository_impl.go
package postgres

import (
	"context"

	"github.com/farmanexo/price-service/internal/domain/entities"
	"github.com/farmanexo/price-service/internal/domain/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PriceHistoryRepositoryImpl implementación PostgreSQL del repositorio de historial de precios
type PriceHistoryRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewPriceHistoryRepository(db *gorm.DB, logger *zap.Logger) *PriceHistoryRepositoryImpl {
	return &PriceHistoryRepositoryImpl{db: db, logger: logger}
}

func (r *PriceHistoryRepositoryImpl) Create(ctx context.Context, history *entities.PriceHistory) error {
	if history.ID == "" {
		history.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *PriceHistoryRepositoryImpl) FindByProductID(ctx context.Context, productID string, limit int) ([]entities.PriceHistory, error) {
	var records []entities.PriceHistory
	if limit <= 0 {
		limit = 50
	}
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("recorded_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *PriceHistoryRepositoryImpl) FindByProductAndPharmacy(ctx context.Context, productID, pharmacyID string, limit int) ([]entities.PriceHistory, error) {
	var records []entities.PriceHistory
	if limit <= 0 {
		limit = 50
	}
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND pharmacy_id = ?", productID, pharmacyID).
		Order("recorded_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *PriceHistoryRepositoryImpl) FindLatestByProduct(ctx context.Context, productID string) ([]entities.PriceHistory, error) {
	var records []entities.PriceHistory
	err := r.db.WithContext(ctx).
		Raw(`SELECT DISTINCT ON (pharmacy_id) * FROM "price".price_history
			 WHERE product_id = ?
			 ORDER BY pharmacy_id, recorded_at DESC`, productID).
		Scan(&records).Error
	return records, err
}

func (r *PriceHistoryRepositoryImpl) FindLatestPrice(ctx context.Context, productID, pharmacyID string) (*entities.PriceHistory, error) {
	var record entities.PriceHistory
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND pharmacy_id = ?", productID, pharmacyID).
		Order("recorded_at DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *PriceHistoryRepositoryImpl) GetAvgPrice(ctx context.Context, productID string) (float64, error) {
	var avg float64
	err := r.db.WithContext(ctx).
		Raw(`SELECT COALESCE(AVG(price), 0) FROM (
			SELECT DISTINCT ON (pharmacy_id) price
			FROM "price".price_history
			WHERE product_id = ?
			ORDER BY pharmacy_id, recorded_at DESC
		) latest_prices`, productID).
		Scan(&avg).Error
	return avg, err
}

func (r *PriceHistoryRepositoryImpl) GetMinPrice(ctx context.Context, productID string) (*entities.PriceHistory, error) {
	var record entities.PriceHistory
	err := r.db.WithContext(ctx).
		Raw(`SELECT * FROM (
			SELECT DISTINCT ON (pharmacy_id) *
			FROM "price".price_history
			WHERE product_id = ?
			ORDER BY pharmacy_id, recorded_at DESC
		) latest_prices ORDER BY price ASC LIMIT 1`, productID).
		Scan(&record).Error
	if err != nil {
		return nil, err
	}
	if record.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &record, nil
}

func (r *PriceHistoryRepositoryImpl) GetMaxPrice(ctx context.Context, productID string) (*entities.PriceHistory, error) {
	var record entities.PriceHistory
	err := r.db.WithContext(ctx).
		Raw(`SELECT * FROM (
			SELECT DISTINCT ON (pharmacy_id) *
			FROM "price".price_history
			WHERE product_id = ?
			ORDER BY pharmacy_id, recorded_at DESC
		) latest_prices ORDER BY price DESC LIMIT 1`, productID).
		Scan(&record).Error
	if err != nil {
		return nil, err
	}
	if record.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &record, nil
}

// Compile-time interface check
var _ repositories.PriceHistoryRepository = (*PriceHistoryRepositoryImpl)(nil)
