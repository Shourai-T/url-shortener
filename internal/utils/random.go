package utils

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Init seed để mỗi lần chạy random không bị giống nhau
func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomString tạo chuỗi ngẫu nhiên độ dài n
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
