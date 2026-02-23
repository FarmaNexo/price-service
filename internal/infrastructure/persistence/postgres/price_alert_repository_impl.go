// internal/infrastructure/persistence/postgres/price_alert_repository_impl.go
package postgres

import (
	"context"

	"github.com/farmanexo/price-service/internal/domain/entities"
	"github.com/farmanexo/price-service/internal/domain/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PriceAlertRepositoryImpl implementación PostgreSQL del repositorio de alertas de precio
type PriceAlertRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewPriceAlertRepository(db *gorm.DB, logger *zap.Logger) *PriceAlertRepositoryImpl {
	return &PriceAlertRepositoryImpl{db: db, logger: logger}
}

func (r *PriceAlertRepositoryImpl) Create(ctx context.Context, alert *entities.PriceAlert) error {
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *PriceAlertRepositoryImpl) FindByID(ctx context.Context, id string) (*entities.PriceAlert, error) {
	var alert entities.PriceAlert
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *PriceAlertRepositoryImpl) FindByUserID(ctx context.Context, userID string) ([]entities.PriceAlert, error) {
	var alerts []entities.PriceAlert
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

func (r *PriceAlertRepositoryImpl) FindActiveByProductID(ctx context.Context, productID string) ([]entities.PriceAlert, error) {
	var alerts []entities.PriceAlert
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND is_active = true", productID).
		Find(&alerts).Error
	return alerts, err
}

func (r *PriceAlertRepositoryImpl) FindByUserAndProduct(ctx context.Context, userID, productID string) (*entities.PriceAlert, error) {
	var alert entities.PriceAlert
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *PriceAlertRepositoryImpl) Update(ctx context.Context, alert *entities.PriceAlert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

func (r *PriceAlertRepositoryImpl) Delete(ctx context.Context, id, userID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&entities.PriceAlert{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *PriceAlertRepositoryImpl) CountActiveByUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.PriceAlert{}).
		Where("user_id = ? AND is_active = true", userID).
		Count(&count).Error
	return count, err
}

// Compile-time interface check
var _ repositories.PriceAlertRepository = (*PriceAlertRepositoryImpl)(nil)
