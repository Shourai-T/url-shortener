package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDatabase khởi tạo kết nối đến Postgres bằng pgxpool
func NewDatabase(dsn string) (*pgxpool.Pool, error) {
	// Parse config
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db config: %w", err)
	}

	// Cấu hình Pool
	config.MaxConns = 25
	config.MinConns = 5
	// ConnMaxLifetime không cần thiết lập thủ công với pgxpool thường (nó tự quản lý tốt)

	// Tạo Pool
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Ping kiểm tra
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	log.Println("✅ Connected to Supabase PostgreSQL successfully via pgxpool")
	return pool, nil
}
