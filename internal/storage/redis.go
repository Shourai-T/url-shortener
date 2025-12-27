package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient khởi tạo Redis client từ host, password, db (legacy/local)
func NewRedisClient(addr string, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisClient{Client: rdb}
}

// NewRedisClientFromURL khởi tạo Redis client từ connection string (Render/Cloud)
// Format: redis://:<password>@<host>:<port>/<db>
func NewRedisClientFromURL(connectionURL string) (*RedisClient, error) {
	opts, err := redis.ParseURL(connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	rdb := redis.NewClient(opts)
	return &RedisClient{Client: rdb}, nil
}

// SetOriginalURL caches the mapping short_code -> original_url with explicit expiration
func (r *RedisClient) SetOriginalURL(ctx context.Context, code string, url string, ttl time.Duration) error {
	return r.Client.Set(ctx, "url:"+code, url, ttl).Err()
}

// GetOriginalURL retrieves original_url from cache
func (r *RedisClient) GetOriginalURL(ctx context.Context, code string) (string, error) {
	return r.Client.Get(ctx, "url:"+code).Result()
}

// IncrementClick increments click count in Redis only (Async Write)
func (r *RedisClient) IncrementClick(ctx context.Context, code string) error {
	// Sử dụng Hash hoặc Set? Đơn giản nhất là dùng string key "click:<code"
	// Tuy nhiên để Worker dễ scan, ta dùng key pattern "click:<code"
	return r.Client.Incr(ctx, "click:"+code).Err()
}

// ScanClickKeys returns all keys matching "click:*"
func (r *RedisClient) ScanClickKeys(ctx context.Context) ([]string, error) {
	var keys []string
	iter := r.Client.Scan(ctx, 0, "click:*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

// GetAndDeleteClick retrieves value and deletes key (Atomic-like for worker)
// Note: Trong môi trường distributed cực lớn cần cẩn thận hơn, nhưng ở đây chấp nhận get rồi del.
// Để an toàn hơn, Worker nên dùng Lua script để Get + Del.
func (r *RedisClient) GetClickCount(ctx context.Context, key string) (int, error) {
	val, err := r.Client.Get(ctx, key).Int()
	return val, err
}

func (r *RedisClient) DeleteKey(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}
