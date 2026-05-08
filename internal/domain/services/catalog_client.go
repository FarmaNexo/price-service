// internal/domain/services/catalog_client.go
package services

import "context"

// ProductInfo información básica de un producto del catálogo
type ProductInfo struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	ActiveIngredient string `json:"active_ingredient"`
	Manufacturer     string `json:"manufacturer"`
	Presentation     string `json:"presentation"`
	IsGeneric        bool   `json:"is_generic"`
	CategoryID       string `json:"category_id"`
	BrandID          string `json:"brand_id"`
}

// CatalogClient interfaz para comunicación con Catalog Service
type CatalogClient interface {
	GetProduct(ctx context.Context, productID string) (*ProductInfo, error)
	// GetProductsByActiveIngredient busca productos con el mismo principio activo (DCI),
	// opcionalmente excluyendo un producto específico. Útil para encontrar alternativas
	// terapéuticas (HU-015 — genéricos vs marca con misma DCI).
	GetProductsByActiveIngredient(ctx context.Context, activeIngredient, excludeID string, limit int) ([]ProductInfo, error)
}
