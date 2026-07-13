-- US4.4: phân loại tài liệu theo nhóm định dạng (pdf/doc/excel/ppt/video).
-- Default rỗng cho các file cũ tạo trước migration này (nếu có).
ALTER TABLE materials ADD COLUMN IF NOT EXISTS file_type TEXT NOT NULL DEFAULT '';
