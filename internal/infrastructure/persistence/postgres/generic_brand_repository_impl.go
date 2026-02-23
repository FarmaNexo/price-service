// internal/infrastructure/persistence/postgres/generic_brand_repository_impl.go
package postgres

import (
	"context"

	"github.com/farmanexo/price-service/internal/domain/entities"
	"github.com/farmanexo/price-service/internal/domain/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GenericBrandRepositoryImpl implementación PostgreSQL del repositorio de comparaciones
type GenericBrandRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewGenericBrandRepository(db *gorm.DB, logger *zap.Logger) *GenericBrandRepositoryImpl {
	return &GenericBrandRepositoryImpl{db: db, logger: logger}
}

func (r *GenericBrandRepositoryImpl) Create(ctx context.Context, comparison *entities.GenericBrandComparison) error {
	if comparison.ID == "" {
		comparison.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(comparison).Error
}

func (r *GenericBrandRepositoryImpl) FindByProductID(ctx context.Context, productID string) ([]entities.GenericBrandComparison, error) {
	var comparisons []entities.GenericBrandComparison
	err := r.db.WithContext(ctx).
		Where("product_id = ? OR related_product_id = ?", productID, productID).
		Order("savings_percentage DESC").
		Find(&comparisons).Error
	return comparisons, err
}

func (r *GenericBrandRepositoryImpl) FindByActiveIngredient(ctx context.Context, activeIngredient string) ([]entities.GenericBrandComparison, error) {
	var comparisons []entities.GenericBrandComparison
	err := r.db.WithContext(ctx).
		Where("active_ingredient = ?", activeIngredient).
		Order("savings_percentage DESC").
		Find(&comparisons).Error
	return comparisons, err
}

func (r *GenericBrandRepositoryImpl) Update(ctx context.Context, comparison *entities.GenericBrandComparison) error {
	return r.db.WithContext(ctx).Save(comparison).Error
}

func (r *GenericBrandRepositoryImpl) Upsert(ctx context.Context, comparison *entities.GenericBrandComparison) error {
	if comparison.ID == "" {
		comparison.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "product_id"}, {Name: "related_product_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"avg_price_product", "avg_price_related", "savings_percentage", "last_calculated_at", "updated_at"}),
		}).
		Create(comparison).Error
}

// Compile-time interface check
var _ repositories.GenericBrandRepository = (*GenericBrandRepositoryImpl)(nil)
