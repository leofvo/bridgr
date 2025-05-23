package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/leofvo/bridgr/internal/config"
	"github.com/leofvo/bridgr/pkg/logger"
)

// RedisStore implements the Store interface using Redis
type RedisStore struct {
	client *redis.Client
	config *config.RedisConfig
}

// NewRedisStore creates a new Redis store instance
func NewRedisStore(cfg *config.RedisConfig) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{
		client: client,
		config: cfg,
	}, nil
}

// HasProcessed checks if an item has been processed by an exporter
func (s *RedisStore) HasProcessed(itemID, exporterID string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("bridgr:processed:%s:%s", exporterID, itemID)

	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if item was processed: %w", err)
	}

	return exists == 1, nil
}

// MarkProcessed marks an item as processed by an exporter
func (s *RedisStore) MarkProcessed(itemID, exporterID string, sourceTTL *time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("bridgr:processed:%s:%s", exporterID, itemID)

	// Use source-specific TTL if provided, otherwise use global TTL
	ttl := s.config.TTL
	if sourceTTL != nil {
		ttl = *sourceTTL
	}

	err := s.client.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to mark item as processed: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// Cleanup removes expired keys
func (s *RedisStore) Cleanup() error {
	ctx := context.Background()
	pattern := "bridgr:processed:*"

	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		ttl, err := s.client.TTL(ctx, key).Result()
		if err != nil {
			logger.Error("Failed to get TTL for key: key=%s error=%v", key, err)
			continue
		}

		if ttl < 0 {
			if err := s.client.Del(ctx, key).Err(); err != nil {
				logger.Error("Failed to delete expired key: key=%s error=%v", key, err)
			}
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	return nil
} 