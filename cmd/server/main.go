package main

import (
	"log"
	"os"

	"github.com/Shourai-T/url-shortener/internal/storage"

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

	log.Println("🚀 Application started. Database connection is ready.")

	// Ở bước sau em sẽ khởi tạo HTTP Server (Gin) ở đây
}
