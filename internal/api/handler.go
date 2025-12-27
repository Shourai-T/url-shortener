package api

import (
	"net/http"

	"github.com/Shourai-T/url-shortener/internal/storage"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *storage.Store
}

func NewHandler(store *storage.Store) *Handler {
	return &Handler{store: store}
}

type ShortenRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
}

func (h *Handler) ShortenURL(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.store.CreateLink(req.OriginalURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to shorten URL"})
		return
	}

	c.JSON(http.StatusOK, ShortenResponse{ShortCode: link.ShortCode})
}
