// internal/infrastructure/cache/cache_service_impl.go
package cache

import (
	"context"
	"time"

	"github.com/farmanexo/price-service/internal/domain/services"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCacheService implementa CacheService usando Redis
type RedisCacheService struct {
	redisClient *RedisClient
	logger      *zap.Logger
}

func NewRedisCacheService(redisClient *RedisClient, logger *zap.Logger) *RedisCacheService {
	return &RedisCacheService{
		redisClient: redisClient,
		logger:      logger,
	}
}

func (c *RedisCacheService) Get(ctx context.Context, key string) (string, error) {
	val, err := c.redisClient.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		c.logger.Warn("Error obteniendo valor de caché",
			zap.String("key", key),
			zap.Error(err),
		)
		return "", err
	}
	return val, nil
}

func (c *RedisCacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := c.redisClient.Client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		c.logger.Warn("Error guardando en caché",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (c *RedisCacheService) Delete(ctx context.Context, key string) error {
	err := c.redisClient.Client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Warn("Error eliminando de caché",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (c *RedisCacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := c.redisClient.Client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.redisClient.Client.Del(ctx, iter.Val()).Err(); err != nil {
			c.logger.Warn("Error eliminando clave por patrón",
				zap.String("key", iter.Val()),
				zap.Error(err),
			)
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}

// Compile-time interface check
var _ services.CacheService = (*RedisCacheService)(nil)
