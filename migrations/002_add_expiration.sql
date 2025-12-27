-- Thêm cột expires_at vào bảng links
ALTER TABLE links 
ADD COLUMN expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '30 days');

-- Tạo index cho expires_at để hỗ trợ việc tìm kiếm/dọn dẹp link hết hạn nhanh hơn
CREATE INDEX idx_links_expires_at ON links(expires_at);
