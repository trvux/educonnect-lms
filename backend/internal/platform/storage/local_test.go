package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"educonnect-lms/backend/internal/platform/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalFileStorage_Save_AvoidsOverwriteOnDuplicateName(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.NewLocalFileStorage(dir)
	require.NoError(t, err)
	ctx := context.Background()

	path1, err := s.Save(ctx, "lesson-1", "slide.pdf", strings.NewReader("noi dung 1"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("lesson-1", "slide.pdf"), path1)

	// Upload lần 2 cùng tên — không được ghi đè file cũ (US4.8).
	path2, err := s.Save(ctx, "lesson-1", "slide.pdf", strings.NewReader("noi dung 2"))
	require.NoError(t, err)
	assert.NotEqual(t, path1, path2, "phải sinh đường dẫn khác nhau khi trùng tên")
	assert.Equal(t, filepath.Join("lesson-1", "slide (1).pdf"), path2)

	path3, err := s.Save(ctx, "lesson-1", "slide.pdf", strings.NewReader("noi dung 3"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("lesson-1", "slide (2).pdf"), path3)

	// Cả 3 file phải còn nguyên trên đĩa với đúng nội dung riêng — chứng
	// minh không file nào bị ghi đè.
	c1, err := os.ReadFile(filepath.Join(dir, path1))
	require.NoError(t, err)
	assert.Equal(t, "noi dung 1", string(c1))

	c2, err := os.ReadFile(filepath.Join(dir, path2))
	require.NoError(t, err)
	assert.Equal(t, "noi dung 2", string(c2))

	c3, err := os.ReadFile(filepath.Join(dir, path3))
	require.NoError(t, err)
	assert.Equal(t, "noi dung 3", string(c3))
}

func TestLocalFileStorage_Delete(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.NewLocalFileStorage(dir)
	require.NoError(t, err)
	ctx := context.Background()

	path, err := s.Save(ctx, "lesson-1", "slide.pdf", strings.NewReader("noi dung"))
	require.NoError(t, err)

	fullPath := filepath.Join(dir, path)
	_, err = os.Stat(fullPath)
	require.NoError(t, err, "file phải tồn tại sau khi Save")

	require.NoError(t, s.Delete(ctx, path))
	_, err = os.Stat(fullPath)
	assert.True(t, os.IsNotExist(err), "file phải bị xóa khỏi đĩa")

	// Xóa lần 2 (file đã không còn) không được coi là lỗi.
	assert.NoError(t, s.Delete(ctx, path))
}
