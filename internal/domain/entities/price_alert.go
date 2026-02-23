// internal/domain/entities/price_alert.go
package entities

import "time"

// PriceAlert representa una alerta de precio configurada por un usuario
type PriceAlert struct {
	ID                    string     `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID                string     `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	ProductID             string     `gorm:"column:product_id;type:uuid;not null" json:"product_id"`
	ProductName           string     `gorm:"column:product_name;type:varchar(500);not null" json:"product_name"`
	TargetPrice           float64    `gorm:"column:target_price;type:decimal(10,2);not null" json:"target_price"`
	CurrentPrice          *float64   `gorm:"column:current_price;type:decimal(10,2)" json:"current_price,omitempty"`
	IsActive              bool       `gorm:"column:is_active;not null;default:true" json:"is_active"`
	IsTriggered           bool       `gorm:"column:is_triggered;not null;default:false" json:"is_triggered"`
	TriggeredAt           *time.Time `gorm:"column:triggered_at" json:"triggered_at,omitempty"`
	TriggeredPharmacyID   *string    `gorm:"column:triggered_pharmacy_id;type:uuid" json:"triggered_pharmacy_id,omitempty"`
	TriggeredPharmacyName *string    `gorm:"column:triggered_pharmacy_name;type:varchar(255)" json:"triggered_pharmacy_name,omitempty"`
	TriggeredPrice        *float64   `gorm:"column:triggered_price;type:decimal(10,2)" json:"triggered_price,omitempty"`
	CreatedAt             time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName retorna el nombre de la tabla con schema
func (PriceAlert) TableName() string {
	return `"price".price_alerts`
}
