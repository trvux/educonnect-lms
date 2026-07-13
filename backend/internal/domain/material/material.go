// Package material là domain của US4.1: Giảng viên tải lên tài liệu/video
// bài giảng (PDF, slide...) gắn với 1 Lesson cụ thể.
package material

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidLessonID = errors.New("material: lesson id là bắt buộc")
	ErrEmptyFileName   = errors.New("material: tên file là bắt buộc")
	ErrEmptyFilePath   = errors.New("material: đường dẫn lưu trữ là bắt buộc")
	ErrNotFound        = errors.New("material: không tìm thấy")
)

// Material đại diện cho 1 file tài liệu (PDF/slide/video) đã upload,
// gắn với 1 Lesson. FilePath là đường dẫn lưu trữ nội bộ (US4.1 dùng local
// disk storage cho bản demo; có thể thay bằng object storage sau này vì
// domain không phụ thuộc trực tiếp implementation lưu trữ — xem
// internal/platform/storage.FileStorage).
type Material struct {
	id         uint
	lessonID   uint
	fileName   string
	filePath   string
	uploadedAt time.Time
}

func NewMaterial(lessonID uint, fileName, filePath string) (*Material, error) {
	if lessonID == 0 {
		return nil, ErrInvalidLessonID
	}
	if fileName == "" {
		return nil, ErrEmptyFileName
	}
	if filePath == "" {
		return nil, ErrEmptyFilePath
	}
	return &Material{lessonID: lessonID, fileName: fileName, filePath: filePath, uploadedAt: time.Now().UTC()}, nil
}

func Rehydrate(id, lessonID uint, fileName, filePath string, uploadedAt time.Time) *Material {
	return &Material{id: id, lessonID: lessonID, fileName: fileName, filePath: filePath, uploadedAt: uploadedAt}
}

func (m *Material) SetID(id uint) { m.id = id }

func (m *Material) ID() uint              { return m.id }
func (m *Material) LessonID() uint        { return m.lessonID }
func (m *Material) FileName() string      { return m.fileName }
func (m *Material) FilePath() string      { return m.filePath }
func (m *Material) UploadedAt() time.Time { return m.uploadedAt }

type Repository interface {
	Create(ctx context.Context, m *Material) error
	FindByID(ctx context.Context, id uint) (*Material, error)
	ListByLesson(ctx context.Context, lessonID uint) ([]*Material, error)
}
