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
