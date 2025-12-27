# Stage 1: Builder
FROM golang:1.20-alpine AS builder

# Cài đặt git và các dependencies cần thiết
RUN apk add --no-cache git

# Thiết lập thư mục làm việc
WORKDIR /app

# Copy go mod và go sum trước để tận dụng Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build ứng dụng
# CGO_ENABLED=0: Build static binary (không phụ thuộc libc), giúp chạy trên Alpine siêu nhẹ
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

# Stage 2: Runner
FROM alpine:latest

# Cài đặt ca-certificates để gọi HTTPS (nếu cần)
RUN apk add --no-cache ca-certificates

WORKDIR /root/

# Copy binary từ builder sang runner
COPY --from=builder /app/server .
# Copy cả .env nếu muốn (nhưng tốt hơn là truyền qua Environment Variable)
COPY --from=builder /app/.env .

# Expose port
EXPOSE 8000

# Chạy ứng dụng
CMD ["./server"]
