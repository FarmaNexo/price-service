// internal/domain/events/price_events.go
package events

import "time"

// PriceEvent representa un evento de precios
type PriceEvent struct {
	EventType  string            `json:"event_type"`
	ProductID  string            `json:"product_id,omitempty"`
	PharmacyID string            `json:"pharmacy_id,omitempty"`
	UserID     string            `json:"user_id,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
	Metadata   map[string]string `json:"metadata"`
}

// Tipos de eventos de precios
const (
	EventPriceAlertTriggered    = "PRICE_ALERT_TRIGGERED"
	EventPriceComparisonCreated = "PRICE_COMPARISON_CREATED"
	EventPriceRecorded          = "PRICE_RECORDED"
)

// NewPriceEvent crea un nuevo evento de precios
func NewPriceEvent(eventType string) PriceEvent {
	return PriceEvent{
		EventType: eventType,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"source":  "price-service",
			"version": "1.0",
		},
	}
}

// WithProduct agrega el ID del producto al evento
func (e PriceEvent) WithProduct(productID string) PriceEvent {
	e.ProductID = productID
	return e
}

// WithPharmacy agrega el ID de la farmacia al evento
func (e PriceEvent) WithPharmacy(pharmacyID string) PriceEvent {
	e.PharmacyID = pharmacyID
	return e
}

// WithUser agrega el ID del usuario al evento
func (e PriceEvent) WithUser(userID string) PriceEvent {
	e.UserID = userID
	return e
}

// InventoryUpdatedEvent evento recibido del Pharmacy Service
type InventoryUpdatedEvent struct {
	EventType  string  `json:"event_type"`
	PharmacyID string  `json:"pharmacy_id"`
	ProductID  string  `json:"product_id"`
	Stock      int     `json:"stock"`
	Price      float64 `json:"price"`
	Timestamp  string  `json:"timestamp"`
	Metadata   struct {
		Source       string `json:"source"`
		Version      string `json:"version"`
		PharmacyName string `json:"pharmacy_name"`
		ProductName  string `json:"product_name"`
	} `json:"metadata"`
}
