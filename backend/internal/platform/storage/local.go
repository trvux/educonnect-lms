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
// tương đối (dùng để lưu vào cột file_path của bảng materials).
func (s *LocalFileStorage) Save(_ context.Context, subdir, fileName string, content io.Reader) (string, error) {
	dir := filepath.Join(s.baseDir, subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("storage: tạo thư mục con lỗi: %w", err)
	}

	relPath := filepath.Join(subdir, fileName)
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
