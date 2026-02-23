// internal/domain/entities/price_history.go
package entities

import "time"

// PriceHistory representa el registro histórico de un precio
type PriceHistory struct {
	ID            string    `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	ProductID     string    `gorm:"column:product_id;type:uuid;not null" json:"product_id"`
	PharmacyID    string    `gorm:"column:pharmacy_id;type:uuid;not null" json:"pharmacy_id"`
	PharmacyName  string    `gorm:"column:pharmacy_name;type:varchar(255);not null" json:"pharmacy_name"`
	ProductName   string    `gorm:"column:product_name;type:varchar(500);not null" json:"product_name"`
	Price         float64   `gorm:"column:price;type:decimal(10,2);not null" json:"price"`
	PreviousPrice *float64  `gorm:"column:previous_price;type:decimal(10,2)" json:"previous_price,omitempty"`
	Currency      string    `gorm:"column:currency;type:varchar(3);not null;default:PEN" json:"currency"`
	Source        string    `gorm:"column:source;type:varchar(50);not null;default:inventory_update" json:"source"`
	RecordedAt    time.Time `gorm:"column:recorded_at;not null" json:"recorded_at"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName retorna el nombre de la tabla con schema
func (PriceHistory) TableName() string {
	return `"price".price_history`
}
