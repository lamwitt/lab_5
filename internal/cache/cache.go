package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"books-api/internal/config"
)

type CacheService struct {
	client *redis.Client
}

func NewCacheService(cfg *config.Config) *CacheService {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Printf("Warning: Redis unavailable: %v. Caching disabled.", err)
		return &CacheService{client: nil}
	}
	log.Println("Redis connected successfully")
	return &CacheService{client: client}
}

func (c *CacheService) available() bool {
	return c.client != nil
}

func (c *CacheService) Get(ctx context.Context, key string) (string, error) {
	if !c.available() {
		return "", redis.Nil
	}
	return c.client.Get(ctx, key).Result()
}

func (c *CacheService) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if !c.available() {
		return nil
	}
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *CacheService) Del(ctx context.Context, keys ...string) error {
	if !c.available() {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// DelByPattern удаляет все ключи, соответствующие паттерну (использует SCAN для безопасного обхода)
func (c *CacheService) DelByPattern(ctx context.Context, pattern string) error {
	if !c.available() {
		return nil
	}
	var cursor uint64
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

func (c *CacheService) Exists(ctx context.Context, key string) bool {
	if !c.available() {
		return false
	}
	n, err := c.client.Exists(ctx, key).Result()
	return err == nil && n > 0
}
