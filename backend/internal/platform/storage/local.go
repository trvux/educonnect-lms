// Package storage cung cấp nơi lưu trữ file vật lý cho US4.1 (upload tài
// liệu bài giảng). Bản demo dùng ổ đĩa local; vì service chỉ phụ thuộc
// interface FileStorage (định nghĩa ở service/material) nên sau này có thể
// đổi sang S3/Cloud Storage mà không sửa domain/service.
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalFileStorage struct {
	baseDir string
}

func NewLocalFileStorage(baseDir string) (*LocalFileStorage, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("storage: tạo thư mục lưu trữ lỗi: %w", err)
	}
	return &LocalFileStorage{baseDir: baseDir}, nil
}

// Save lưu file vào <baseDir>/<subdir>/<fileName> và trả về đường dẫn
// tương đối (dùng để lưu vào cột file_path của bảng materials). US4.8:
// nếu đã tồn tại file trùng tên trong cùng subdir, tự đổi tên file lưu trên
// đĩa theo kiểu "ten (1).pdf", "ten (2).pdf"... (giống Google Drive) thay
// vì os.Create ghi đè âm thầm file cũ — tên hiển thị cho người dùng
// (Material.FileName, do caller truyền nguyên vẹn) không đổi, chỉ đường dẫn
// lưu trữ thật trên đĩa được làm khác nhau.
func (s *LocalFileStorage) Save(_ context.Context, subdir, fileName string, content io.Reader) (string, error) {
	dir := filepath.Join(s.baseDir, subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("storage: tạo thư mục con lỗi: %w", err)
	}

	storedName := uniqueFileName(dir, fileName)
	relPath := filepath.Join(subdir, storedName)
	fullPath := filepath.Join(s.baseDir, relPath)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("storage: tạo file lỗi: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, content); err != nil {
		return "", fmt.Errorf("storage: ghi file lỗi: %w", err)
	}
	return relPath, nil
}

// Delete xóa file vật lý tại đường dẫn tương đối path (US4.8). Không coi
// "file đã không tồn tại từ trước" là lỗi — mục tiêu cuối cùng (file không
// còn trên đĩa) đã đạt được.
func (s *LocalFileStorage) Delete(_ context.Context, path string) error {
	fullPath := filepath.Join(s.baseDir, path)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: xóa file lỗi: %w", err)
	}
	return nil
}

// uniqueFileName trả về 1 tên file không trùng với file nào đang có trong
// dir, chèn thêm hậu tố " (n)" trước phần mở rộng nếu cần.
func uniqueFileName(dir, fileName string) string {
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	candidate := fileName
	for i := 1; ; i++ {
		if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s (%d)%s", base, i, ext)
	}
}
