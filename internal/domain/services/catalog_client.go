// internal/domain/services/catalog_client.go
package services

import "context"

// ProductInfo información básica de un producto del catálogo
type ProductInfo struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ActiveIngredient string `json:"active_ingredient"`
	Presentation     string `json:"presentation"`
	IsGeneric        bool   `json:"is_generic"`
	CategoryID       string `json:"category_id"`
	BrandID          string `json:"brand_id"`
}

// CatalogClient interfaz para comunicación con Catalog Service
type CatalogClient interface {
	GetProduct(ctx context.Context, productID string) (*ProductInfo, error)
	GetProductsByActiveIngredient(ctx context.Context, activeIngredient string) ([]ProductInfo, error)
}
