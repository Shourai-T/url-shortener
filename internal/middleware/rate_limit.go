package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimiterMiddleware trả về gin middleware để giới hạn request
// rate: chuỗi định dạng (vd: "10-M" -> 10 req/phút, "5-S" -> 5 req/giây)
func RateLimiterMiddleware(rateString string) gin.HandlerFunc {
	// 1. Parse rate string
	rate, err := limiter.NewRateFromFormatted(rateString)
	if err != nil {
		panic(fmt.Sprintf("Invalid rate string: %v", err))
	}

	// 2. Tạo in-memory store
	store := memory.NewStore()

	// 3. Tạo instance limiter
	instance := limiter.New(store, rate)

	// 4. Tạo middleware cho Gin
	return mgin.NewMiddleware(instance, mgin.WithLimitReachedHandler(func(c *gin.Context) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "Too many requests. Limit: " + rateString,
			"retry_after": time.Now().Add(time.Minute).Format(time.RFC3339),
		})
	}))
}
