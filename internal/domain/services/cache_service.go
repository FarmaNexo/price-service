// internal/domain/services/cache_service.go
package services

import (
	"context"
	"time"
)

// CacheService interfaz para servicio de caché
type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
}
