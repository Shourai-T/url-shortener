package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Shourai-T/url-shortener/internal/model"
	"github.com/Shourai-T/url-shortener/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db    *pgxpool.Pool
	redis *RedisClient
}

func NewStore(db *pgxpool.Pool, redis *RedisClient) *Store {
	return &Store{db: db, redis: redis}
}

// CreateLink thực hiện sinh mã và lưu vào DB
func (s *Store) CreateLink(originalURL string) (*model.Link, error) {
	var link model.Link
	link.OriginalURL = originalURL
	ctx := context.Background()

	// Thử tối đa 3 lần nếu bị trùng mã code (Collision handling)
	for i := 0; i < 3; i++ {
		link.ShortCode = utils.GenerateRandomString(6) // Sinh mã 6 ký tự

		query := `INSERT INTO links (original_url, short_code) 
                  VALUES ($1, $2) RETURNING id`

		// Sử dụng pgxpool
		err := s.db.QueryRow(ctx, query, link.OriginalURL, link.ShortCode).Scan(&link.ID)
		if err == nil {
			// Thành công, cache ngay vào Redis để đọc nhanh
			_ = s.redis.SetOriginalURL(ctx, link.ShortCode, link.OriginalURL)
			return &link, nil
		}

		// Nếu lỗi không phải do trùng lặp thì return lỗi luôn
		fmt.Printf("Retry generating code due to error: %v\n", err)
	}

	return nil, fmt.Errorf("failed to generate unique code after retries")
}

// GetAndIncrement lấy URL gốc và tăng lượt click (Async via Redis)
func (s *Store) GetAndIncrement(shortCode string) (string, error) {
	ctx := context.Background()

	// 1. Kiểm tra Cache Redis trước (Read Path)
	cachedURL, err := s.redis.GetOriginalURL(ctx, shortCode)
	if err == nil && cachedURL != "" {
		// Cache Hit
		// 2. Async Write: Tăng click trong Redis, không ghi DB ngay
		go func() {
			_ = s.redis.IncrementClick(context.Background(), shortCode)
		}()
		return cachedURL, nil
	}

	// 3. Cache Miss: Đọc từ DB
	query := `SELECT original_url FROM links WHERE short_code = $1`
	var originalURL string
	err = s.db.QueryRow(ctx, query, shortCode).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("link not found")
		}
		return "", fmt.Errorf("failed to get link: %w", err)
	}

	// 4. Update Cache lại cho lần sau
	_ = s.redis.SetOriginalURL(ctx, shortCode, originalURL)

	// 5. Tăng click (Async)
	go func() {
		_ = s.redis.IncrementClick(context.Background(), shortCode)
	}()

	return originalURL, nil
}

// GetLinkStats để xem thông tin link
func (s *Store) GetLinkStats(shortCode string) (*model.Link, error) {
	var link model.Link
	ctx := context.Background()
	query := `SELECT original_url, short_code, click_count, created_at FROM links WHERE short_code = $1`
	err := s.db.QueryRow(ctx, query, shortCode).Scan(&link.OriginalURL, &link.ShortCode, &link.ClickCount, &link.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// GetAllLinks lấy danh sách link có phân trang
func (s *Store) GetAllLinks(limit, offset int) ([]model.Link, error) {
	query := `SELECT original_url, short_code, click_count, created_at 
	          FROM links 
	          ORDER BY created_at DESC 
	          LIMIT $1 OFFSET $2`

	ctx := context.Background()
	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.Link
	for rows.Next() {
		var link model.Link
		if err := rows.Scan(&link.OriginalURL, &link.ShortCode, &link.ClickCount, &link.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}
