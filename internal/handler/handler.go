package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Shourai-T/url-shortener/internal/model"
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

func (h *Handler) RedirectHandler(c *gin.Context) {
	code := c.Param("code")

	originalURL, err := h.store.GetAndIncrement(code)
	if err != nil {
		log.Printf("Error in RedirectHandler: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Short link not found"})
		return
	}

	// Redirect 302 (Found)
	c.Redirect(http.StatusFound, originalURL)
}

// GetStatsHandler: Xem thống kê
func (h *Handler) GetStats(c *gin.Context) {
	code := c.Param("code")
	link, err := h.store.GetLinkStats(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}
	c.JSON(http.StatusOK, link)
}

// ListLinks: Lấy danh sách link (Pagination)
func (h *Handler) ListLinks(c *gin.Context) {
	// Lấy tham số page va limit từ URL query
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	} // Hard limit để tránh user query quá nhiều

	offset := (page - 1) * limit

	links, err := h.store.GetAllLinks(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
		return
	}

	// Nếu danh sách rỗng thì trả về mảng rỗng thay vì null
	if links == nil {
		links = []model.Link{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  links,
		"page":  page,
		"limit": limit,
	})
}
