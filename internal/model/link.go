package model

import "time"

type Link struct {
	ID          int       `json:"-"`                      // ID nội bộ, không expose ra ngoài
	OriginalURL string    `json:"url" binding:"required"` // Input bắt buộc
	ShortCode   string    `json:"short_code"`
	ClickCount  int       `json:"clicks"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"` // Thời gian hết hạn
}
