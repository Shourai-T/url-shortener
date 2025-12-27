package model

type Link struct {
	ID          int    `json:"-"`                      // ID nội bộ, không expose ra ngoài
	OriginalURL string `json:"url" binding:"required"` // Input bắt buộc
	ShortCode   string `json:"short_code"`
	ClickCount  int    `json:"clicks"`
}
