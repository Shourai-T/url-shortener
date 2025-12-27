# Go URL Shortener Service

Đây là bài test kỹ thuật cho vị trí **Golang Intern**. Dự án xây dựng một dịch vụ rút gọn liên kết (tương tự Bit.ly) với hiệu năng cao, xử lý Concurrency an toàn và cấu trúc Clean Architecture.

## 🚀 Tính năng đã hoàn thành

1.  **Shorten URL**: Rút gọn link dài thành mã 6 ký tự (Base62).
2.  **Redirect**: Chuyển hướng người dùng về link gốc khi truy cập link ngắn.
3.  **Click Tracking**: Đếm số lượt click (View count).
4.  **Concurrency Safe**: Đảm bảo bộ đếm click chính xác tuyệt đối ngay cả khi có hàng nghìn request cùng lúc.
5.  **Link Stats**: Xem thông tin chi tiết của link (URL gốc, ngày tạo, số click).

## 🛠 Tech Stack

- **Language**: Golang 1.20+
- **Framework**: Gin Gonic (High performance HTTP web framework)
- **Database**: PostgreSQL (Supabase)
- **Driver**: pgx/v5 (Driver thuần Go hiệu năng cao cho Postgres)
- **Architecture**: Layered Architecture (Handler -> Store -> Database)

## ⚙️ Cài đặt & Chạy dự án

### 1. Prerequisites

- Go (Golang) đã cài đặt.
- Database PostgreSQL (hoặc Supabase account).

### 2. Setup

Clone dự án và cài đặt dependencies:

```bash
git clone https://github.com/Shourai-T/url-shortener.git
cd url-shortener
go mod tidy
```

### 3. Cấu hình

Tạo file `.env` tại thư mục gốc và điền Connection String của Database:

```env
# Thay thế [YOUR-PASSWORD] bằng mật khẩu thực
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.epecmzghqxnteadhgkls.supabase.co:5432/postgres

```

### 4. Database Migration

Chạy script SQL sau trong SQL Editor của Supabase để tạo bảng:

```sql
CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    short_code VARCHAR(10) NOT NULL UNIQUE,
    click_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index giúp tìm kiếm link ngắn cực nhanh (O(log N))
CREATE INDEX idx_short_code ON links(short_code);

```

### 5. Chạy Server

```bash
go run cmd/server/main.go

```

Server sẽ khởi động tại `http://localhost:8000`.

---

## 📡 API Documentation

### 1. Tạo Link Rút Gọn

- **Endpoint**: `POST /shorten`
- **Body**:

```json
{
  "original_url": "https://google.com"
}
```

- **Response**:

```json
{
  "short_code": "DjXhpM"
}
```

### 2. Truy cập Link (Redirect)

- **Endpoint**: `GET /:code` (Ví dụ: `http://localhost:8000/Ab3d9Z`)
- **Behavior**: Server trả về HTTP 302 Found và chuyển hướng trình duyệt sang URL gốc. Đồng thời tăng `click_count` lên 1.

### 3. Xem Thống Kê

- **Endpoint**: `GET /api/stats/:code`
- **Response**:

```json
{
  "url": "https://www.google.com",
  "short_code": "Ab3d9Z",
  "click_count": 42,
  "created_at": "2023-10-27T10:00:00Z"
}
```

---

## 🧠 Thiết kế & Quyết định Kỹ thuật (Design Decisions)

### 1. Tại sao chọn cấu trúc project này?

Tôi sử dụng **Standard Go Project Layout** với thư mục `internal` để đóng gói logic:

- `cmd/server`: Entry point, giữ cho `main` gọn gàng.
- `internal/api`: Xử lý HTTP Request/Response (Presentation Layer).
- `internal/storage`: Xử lý Database logic (Data Layer).
- **Lợi ích**: Dễ dàng mở rộng, viết Unit Test và bảo trì.

### 2. Thuật toán sinh mã (Shortening Algorithm)

Tôi sử dụng phương pháp **Random String Base62** (`a-z`, `A-Z`, `0-9`).

- **Không gian mẫu**: Với độ dài 6 ký tự, có tỷ tổ hợp. Đủ lớn để tránh trùng lặp trong thời gian dài.
- **Collision Handling**: Mặc dù xác suất thấp, tôi vẫn xử lý trường hợp trùng mã bằng cơ chế **Retry** (thử lại tối đa 3 lần) nếu DB báo lỗi Duplicate Key.

### 3. Giải quyết vấn đề Concurrency (Race Condition)

Đây là thách thức lớn nhất: Nếu 1000 users click cùng lúc, việc đọc `click_count` lên rồi cộng 1 ở code Go sẽ gây sai lệch dữ liệu.

**Giải pháp**: Tôi sử dụng **Atomic Update** ở mức Database.

```sql
UPDATE links
SET click_count = click_count + 1
WHERE short_code = $1
RETURNING original_url

```

- PostgreSQL sẽ lock row đó lại và thực hiện update tuần tự.
- Đảm bảo tính **ACID** và dữ liệu luôn chính xác 100%.

---
