package worker

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Shourai-T/url-shortener/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SyncWorker struct {
	db    *pgxpool.Pool
	redis *storage.RedisClient
}

func NewSyncWorker(db *pgxpool.Pool, redis *storage.RedisClient) *SyncWorker {
	return &SyncWorker{db: db, redis: redis}
}

// Start chạy worker định kỳ
func (w *SyncWorker) Start() {
	ticker := time.NewTicker(10 * time.Second) // Sync mỗi 10s
	go func() {
		for {
			<-ticker.C
			w.SyncClicks()
		}
	}()
}

// SyncClicks lấy click count từ Redis và update vào Postgres
func (w *SyncWorker) SyncClicks() {
	ctx := context.Background()
	keys, err := w.redis.ScanClickKeys(ctx)
	if err != nil {
		log.Printf("[Worker] Failed to scan keys: %v", err)
		return
	}

	if len(keys) == 0 {
		return
	}

	log.Printf("[Worker] Syncing %d keys...", len(keys))

	for _, key := range keys {
		// Key format: "click:<code>"
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			continue
		}
		shortCode := parts[1]

		count, err := w.redis.GetClickCount(ctx, key)
		if err != nil {
			continue
		}

		if count > 0 {
			// Update DB
			_, err := w.db.Exec(ctx, "UPDATE links SET click_count = click_count + $1 WHERE short_code = $2", count, shortCode)
			if err == nil {
				// Thành công thì xóa key trong Redis để reset bộ đếm
				_ = w.redis.DeleteKey(ctx, key)
			} else {
				log.Printf("[Worker] Failed to update DB for %s: %v", shortCode, err)
			}
		}
	}
}
