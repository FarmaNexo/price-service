// internal/domain/entities/generic_brand_comparison.go
package entities

import "time"

// GenericBrandComparison representa una comparación entre genérico y marca
type GenericBrandComparison struct {
	ID                 string    `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	ProductID          string    `gorm:"column:product_id;type:uuid;not null" json:"product_id"`
	ProductName        string    `gorm:"column:product_name;type:varchar(500);not null" json:"product_name"`
	ActiveIngredient   string    `gorm:"column:active_ingredient;type:varchar(255);not null" json:"active_ingredient"`
	IsGeneric          bool      `gorm:"column:is_generic;not null;default:false" json:"is_generic"`
	RelatedProductID   string    `gorm:"column:related_product_id;type:uuid;not null" json:"related_product_id"`
	RelatedProductName string    `gorm:"column:related_product_name;type:varchar(500);not null" json:"related_product_name"`
	RelatedIsGeneric   bool      `gorm:"column:related_is_generic;not null;default:false" json:"related_is_generic"`
	AvgPriceProduct    float64   `gorm:"column:avg_price_product;type:decimal(10,2);not null;default:0" json:"avg_price_product"`
	AvgPriceRelated    float64   `gorm:"column:avg_price_related;type:decimal(10,2);not null;default:0" json:"avg_price_related"`
	SavingsPercentage  float64   `gorm:"column:savings_percentage;type:decimal(5,2);not null;default:0" json:"savings_percentage"`
	LastCalculatedAt   time.Time `gorm:"column:last_calculated_at;not null" json:"last_calculated_at"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName retorna el nombre de la tabla con schema
func (GenericBrandComparison) TableName() string {
	return `"price".generic_brand_comparisons`
}
