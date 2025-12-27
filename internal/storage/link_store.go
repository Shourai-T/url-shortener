package storage

import (
	"database/sql"
	"fmt"

	"github.com/Shourai-T/url-shortener/internal/model"
	"github.com/Shourai-T/url-shortener/internal/utils"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

// CreateLink thực hiện sinh mã và lưu vào DB
func (s *Store) CreateLink(originalURL string) (*model.Link, error) {
	var link model.Link
	link.OriginalURL = originalURL

	// Thử tối đa 3 lần nếu bị trùng mã code (Collision handling)
	for i := 0; i < 3; i++ {
		link.ShortCode = utils.GenerateRandomString(6) // Sinh mã 6 ký tự

		query := `INSERT INTO links (original_url, short_code) 
                  VALUES ($1, $2) RETURNING id`

		err := s.DB.QueryRow(query, link.OriginalURL, link.ShortCode).Scan(&link.ID)
		if err == nil {
			// Thành công
			return &link, nil
		}

		// Nếu lỗi không phải do trùng lặp thì return lỗi luôn
		// (Trong thực tế nên check err code của Postgres xem có phải unique violation không)
		fmt.Printf("Retry generating code due to error: %v\n", err)
	}

	return nil, fmt.Errorf("failed to generate unique code after retries")
}

// GetAndIncrement lấy URL gốc và tăng lượt click (Atomic Update)
func (s *Store) GetAndIncrement(shortCode string) (string, error) {
	query := `UPDATE links 
	          SET click_count = click_count + 1 
	          WHERE short_code = $1 
	          RETURNING original_url`

	var originalURL string
	err := s.DB.QueryRow(query, shortCode).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("link not found")
		}
		return "", fmt.Errorf("failed to update click count: %w", err)
	}

	return originalURL, nil
}

// GetLinkStats để xem thông tin link
func (s *Store) GetLinkStats(shortCode string) (*model.Link, error) {
	var link model.Link
	query := `SELECT original_url, short_code, click_count FROM links WHERE short_code = $1`
	err := s.DB.QueryRow(query, shortCode).Scan(&link.OriginalURL, &link.ShortCode, &link.ClickCount)
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

	rows, err := s.DB.Query(query, limit, offset)
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

	// Kiểm tra lỗi sau khi lặp
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}
