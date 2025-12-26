-- Tạo bảng links
CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,              -- ID tự tăng nội bộ
    original_url TEXT NOT NULL,         -- URL gốc
    short_code VARCHAR(10) NOT NULL,    -- Mã rút gọn (ví dụ: aZb12)
    click_count INT DEFAULT 0,          -- Đếm số lượt click
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() -- Thời gian tạo
);

-- Tạo Index cho short_code để tìm kiếm cực nhanh (O(log N))
CREATE UNIQUE INDEX idx_links_short_code ON links(short_code);