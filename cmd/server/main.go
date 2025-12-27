package main

import (
	"log"
	"os"

	"github.com/Shourai-T/url-shortener/internal/handler"
	"github.com/Shourai-T/url-shortener/internal/storage"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load biến môi trường
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system env vars")
	}

	// 2. Lấy Database URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// 3. Khởi tạo kết nối DB
	db, err := storage.NewDatabase(dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close() // Đóng kết nối khi app dừng

	log.Println("Application started. Database connection is ready.")

	// 4. Initialize Dependency
	store := storage.NewStore(db)
	handler := handler.NewHandler(store)

	// 5. Setup Router
	r := gin.Default()
	r.POST("/shorten", handler.ShortenURL)
	r.GET("/:code", handler.RedirectHandler)

	api := r.Group("/api")
	api.GET("/stats/:code", handler.GetStats)
	api.GET("/links", handler.ListLinks)

	// 6. Run Server
	log.Println("Running on :8000")
	if err := r.Run(":8000"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
