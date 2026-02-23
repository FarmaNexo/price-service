// internal/domain/services/pharmacy_client.go
package services

import "context"

// PharmacyInventoryItem representa un item del inventario de farmacia
type PharmacyInventoryItem struct {
	PharmacyID   string  `json:"pharmacy_id"`
	PharmacyName string  `json:"pharmacy_name"`
	ProductID    string  `json:"product_id"`
	Stock        int     `json:"stock"`
	Price        float64 `json:"price"`
	IsAvailable  bool    `json:"is_available"`
}

// PharmacyClient interfaz para comunicación con Pharmacy Service
type PharmacyClient interface {
	GetProductPrices(ctx context.Context, productID string) ([]PharmacyInventoryItem, error)
	GetPharmacyInfo(ctx context.Context, pharmacyID string) (*PharmacyInfo, error)
}

// PharmacyInfo información básica de una farmacia
type PharmacyInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
