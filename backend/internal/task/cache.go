package task

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	listKeyPrefix = "tasks:list:"
	listCacheTTL  = 60 * time.Second
)

type ListCache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	InvalidateLists(ctx context.Context) error
}

type RedisListCache struct {
	rdb *redis.Client
}

func NewRedisListCache(rdb *redis.Client) *RedisListCache {
	return &RedisListCache{rdb: rdb}
}

func (c *RedisListCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

func (c *RedisListCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, val, ttl).Err()
}

// InvalidateLists uses SCAN rather than KEYS so it never blocks Redis.
func (c *RedisListCache) InvalidateLists(ctx context.Context) error {
	var cursor uint64
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, listKeyPrefix+"*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if next == 0 {
			return nil
		}
		cursor = next
	}
}
