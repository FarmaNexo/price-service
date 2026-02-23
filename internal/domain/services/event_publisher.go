// internal/domain/services/event_publisher.go
package services

import (
	"context"

	"github.com/farmanexo/price-service/internal/domain/events"
)

// EventPublisher interfaz para publicar eventos de precios
type EventPublisher interface {
	Publish(ctx context.Context, event events.PriceEvent) error
}
