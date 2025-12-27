# Go URL Shortener Service

Đây là bài test kỹ thuật cho vị trí **Golang Intern**. Dự án xây dựng một dịch vụ rút gọn liên kết (tương tự Bit.ly) với hiệu năng cao, xử lý Concurrency an toàn và cấu trúc Clean Architecture.

## 🚀 Tính năng đã hoàn thành

1.  **Shorten URL**: Rút gọn link dài thành mã 6 ký tự (Base62).
2.  **Redirect**: Chuyển hướng người dùng về link gốc khi truy cập link ngắn.
3.  **Click Tracking**: Đếm số lượt click (View count).
4.  **Concurrency Safe**: Đảm bảo bộ đếm click chính xác tuyệt đối ngay cả khi có hàng nghìn request cùng lúc.
5.  **Link Stats**: Xem thông tin chi tiết của link (URL gốc, ngày tạo, số click, ngày hết hạn).
6.  **Rate Limiting**: Giới hạn mỗi IP chỉ được tạo 10 link/phút để chống Spam.
7.  **High Performance**: Sử dụng Redis Caching và Async Queue để xử lý hàng nghìn request mỗi giây.

## 🛠 Tech Stack

- **Language**: Golang 1.20+
- **Framework**: Gin Gonic (High performance HTTP web framework)
- **Framework**: Gin Gonic (High performance HTTP web framework)
- **Database**: PostgreSQL (Supabase) + Redis (Caching)
- **Driver**: pgxpool/v5 (Connection Pooling)
- **Container**: Docker + Docker Compose
- **Architecture**: Layered Architecture (Handler -> Store -> Database)

## ⚙️ Cài đặt & Chạy dự án

### 1. Prerequisites

- Go (Golang) đã cài đặt.
- Database PostgreSQL (hoặc Supabase account).
- Redis (cho Caching & Async Queue).

### 2. Setup

**Bước 1: Khởi động Redis (bắt buộc)**

Nếu dùng Docker:

```bash
docker run -d --name redis-stack -p 6379:6379 -p 8001:8001 redis/redis-stack:latest
```

_Lưu ý: Image `redis-stack` đi kèm với giao diện quản lý trực quan tại `http://localhost:8001`._

**Bước 2: Clone dự án**

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

### 6. Chạy với Docker (Khuyên dùng)

Chỉ cần một lệnh duy nhất để dựng toàn bộ stack:

```bash
docker-compose up -d --build
```

Hệ thống bao gồm:

- **App**: `http://localhost:8000`
- **RedisUI**: `http://localhost:8001`

### 7. Link bài viết hướng dẫn Deploy (Render.com)

Để deploy dự án này lên Render:

1.  Fork repo này về GitHub của bạn.
2.  Truy cập [Render Dashboard](https://dashboard.render.com/blueprints).
3.  Chọn **New Blueprint Instance**.
4.  Kết nối với repo GitHub vừa Fork.
5.  Render sẽ tự động đọc file `render.yaml` và dựng lên Web Service + Redis + Database.
6.  Ngồi chờ 2 phút và tận hưởng kết quả!

---

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

### 4. Liệt Kê Danh Sách (Management API)

- **Endpoint**: `GET /api/links`
- **Mục đích**: Xem danh sách các link đã tạo, hỗ trợ phân trang để tối ưu hiệu năng.
- **Query Params**:
  - `page`: Số trang cần xem (Mặc định: 1).
  - `limit`: Số lượng link trên mỗi trang (Mặc định: 10, Tối đa: 100).
- **Ví dụ**: `GET /api/links?page=1&limit=5`
- **Response**:
  ```json
  {
    "data": [
      {
        "url": "https://google.com",
        "short_code": "aBc123",
        "click_count": 15,
        "created_at": "2023-12-28T10:00:00Z"
      }
      // ... các link khác
    ],
    "page": 1,
    "limit": 5
  }
  ```

---

## 🧠 Thiết kế & Quyết định Kỹ thuật (Design Decisions)

### 1. Tại sao chọn PostgreSQL?

Em sử dụng **PostgreSQL** vì:

- Yêu cầu bài toán cần đếm click_count chính xác. Postgres hỗ trợ ACID, giúp đảm bảo dữ liệu không bị sai lệch khi có concurrency (điều mà NoSQL như MongoDB cần xử lý phức tạp hơn).
- Supabase được chọn vì khả năng setup nhanh chóng và hạ tầng ổn định.

### 2. Thuật toán sinh mã (Shortening Algorithm)

Em sử dụng phương pháp **Random String Base62** (`a-z`, `A-Z`, `0-9`).

- **Không gian mẫu**: Với độ dài 6 ký tự, có tỷ tổ hợp. Đủ lớn để tránh trùng lặp trong thời gian dài.
- **Collision Handling**: Mặc dù xác suất thấp, em vẫn xử lý trường hợp trùng mã bằng cơ chế **Retry** (thử lại tối đa 3 lần) nếu DB báo lỗi Duplicate Key.

### 3. Giải quyết vấn đề Concurrency (Race Condition)

Đây là thách thức lớn nhất: Nếu 1000 users click cùng lúc, việc đọc `click_count` lên rồi cộng 1 ở code Go sẽ gây sai lệch dữ liệu.

**Giải pháp**: Em sử dụng **Atomic Update** ở mức Database.

```sql
UPDATE links
SET click_count = click_count + 1
WHERE short_code = $1
RETURNING original_url

```

- PostgreSQL sẽ lock row đó lại và thực hiện update tuần tự.
- Đảm bảo tính **ACID** và dữ liệu luôn chính xác 100%.

### 4. Chiến lược Performance (Pagination)

- **Vấn đề**: Đề bài đặt ra thách thức "Nếu có 1 triệu links thì query ra sao?". Việc query toàn bộ (`SELECT *`) sẽ gây quá tải Database và tràn bộ nhớ (OOM) Application Server.
- **Giải pháp**: Em triển khai **Offset-based Pagination** (Phân trang).
- **Implementation**: Sử dụng câu lệnh `LIMIT $1 OFFSET $2` trong PostgreSQL.
- **Kết quả**: API luôn phản hồi nhanh (low latency) và tiêu tốn ít RAM, bất kể kích thước dữ liệu trong bảng `links` lớn đến đâu.

### 5. Tại sao chọn RESTful API (thay vì GraphQL/gRPC)?

- **Lý do**: Em chọn REST thay vì gRPC hay GraphQL vì tính đơn giản, dễ debug và tận dụng được khả năng caching của HTTP protocol cho các request Redirect.
- **Format**: JSON chuẩn (snake_case) dễ dàng tích hợp với Frontend.

## Trade-offs (Đánh đổi)

Trong quá trình làm, em đã phải cân nhắc giữa các lựa chọn:

### 1. SQL Driver (pgx) vs ORM (Gorm)

#### Em chọn pgx (SQL thuần) thay vì Gorm.

**Lý do:** Mặc dù ORM code ngắn hơn, nhưng dùng SQL thuần giúp em kiểm soát tối đa câu query, đặc biệt là tính năng RETURNING và Atomic Update để tối ưu hiệu năng. Đây cũng là cách để em rèn luyện kỹ năng SQL.

### 2. Random String vs Auto-Increment ID (Base62 Conversion)

#### Em chọn Random String thay vì Convert ID sang Base62.

**Lý do:** Cách Random giúp URL khó đoán hơn (Security), người ngoài không biết hệ thống có bao nhiêu link.

**Nhược điểm:** Phải xử lý vấn đề trùng mã (Collision), nhưng với không gian mẫu 56 tỷ thì tỷ lệ trùng cực thấp, chấp nhận được.

### 3. Pagination (Offset) vs Cursor-based

#### Em chọn Offset Pagination (LIMIT, OFFSET).

**Lý do:** Dễ cài đặt, phù hợp với UI trang số truyền thống.

**Nhược điểm:** Sẽ chậm nếu offset quá lớn (ví dụ trang 1 triệu), nhưng với yêu cầu hiện tại thì đây là giải pháp cân bằng tốt nhất.

## 🛑 Challenges & Limitations (Self-Review)

### 1. Validation & Edge Cases

- **Hiện tại**: Hệ thống chỉ kiểm tra URL không rỗng.
- **Vấn đề**: Người dùng có thể nhập chuỗi không phải URL (ví dụ: "hello world") hoặc Local IP gây lỗi SSRF.
- **Giải pháp (Future)**: Sử dụng package `net/url` để `ParseRequestURI` và kiểm tra scheme (http/https).

### 2. Scalability (Traffic x100)

- **Vấn đề**: Khi traffic tăng đột biến, Database sẽ là nút thắt cổ chai (Bottleneck) vì mọi request redirect đều phải đọc/ghi DB.
- **Giải pháp**:
  - **Caching**: Sử dụng **Redis** để lưu cặp key-value `short_code -> original_url`. Request đọc sẽ hit Cache trước (tốc độ < 5ms), chỉ hit DB khi cache miss.
  - **Async Write**: Việc cập nhật `click_count` không cần realtime tức thì. Có thể ghi vào Redis trước, sau đó dùng Worker đồng bộ xuống Postgres sau mỗi 1 phút (Batch Processing).

### 3. Concurrency (Create Link)

- **Câu hỏi**: Nếu 2 request cùng tạo 1 URL thì sao?
- **Quyết định**: Hệ thống hiện tại cho phép tạo 2 mã short code khác nhau cho cùng 1 URL gốc.
- **Lý do**: Hỗ trợ nhu cầu Marketing (A/B Testing). Ví dụ: User muốn tracking riêng link này khi post Facebook và khi gửi Email.

### 4. Security

- **Đã làm**: Chống SQL Injection (Parameterized Query), Chống ID Enumeration (Random String).
- **Cần làm thêm**:
  - **Rate Limiting**: Chặn IP spam tạo hàng loạt link.
  - **Phishing Check**: Tích hợp Google Safe Browsing API để chặn rút gọn link lừa đảo/độc hại.

---

### 6. Scalability Implementation (Redis & Async)

Em đã tích hợp Redis để giải quyết bài toán hiệu năng lớn:

#### 6.1. Caching (Read Path)

- Khi user truy cập Link ngắn, hệ thống kiểm tra Redis trước.
- **Cache Hit**: Trả về URL ngay lập tức (< 1ms), không gọi vào Postgres.
- **Cache Miss**: Đọc từ Postgres -> Lưu vào Redis -> Trả về.
- **Expiration**: Cache tự động hết hạn theo thời gian sống của Link (30 ngày).

#### 6.2. Async Click Counting (Write Path)

- Thay vì UPDATE vào Postgres mỗi lần click (gây lock row, chậm), em dùng Redis `INCR` (Atomic).
- Một **Background Worker** chạy ngầm mỗi 10 giây sẽ gom số lượng click từ Redis và cập nhật xuống Postgres một lần (Batch Update).
- **Kết quả**: Giảm tải Write cho DB đến 99%.

### 7. Rate Limiting

Sử dụng thuật toán **Token Bucket** (via `ulule/limiter`) để giới hạn mỗi IP chỉ được gọi API tạo link 10 lần/phút.

### 8. Link Expiration

- Mặc định mỗi link có tuổi thọ **30 ngày**.
- Hệ thống tự động chặn truy cập nếu quá hạn.
- Redis TTL được đồng bộ với thời gian hết hạn của Link.

---
