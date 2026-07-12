// Package material là service layer của US4.1: Giảng viên tải lên tài
// liệu/video bài giảng cho từng bài học.
package material

import (
	"context"
	"io"
	"strconv"

	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/material"
)

// FileStorage là port lưu trữ file vật lý, implement bởi
// internal/platform/storage (local disk cho bản demo).
type FileStorage interface {
	Save(ctx context.Context, subdir, fileName string, content io.Reader) (path string, err error)
}

type Service struct {
	materials material.Repository
	lessons   curriculum.LessonRepository
	storage   FileStorage
}

func NewService(materials material.Repository, lessons curriculum.LessonRepository, storage FileStorage) *Service {
	return &Service{materials: materials, lessons: lessons, storage: storage}
}

// Upload hiện thực US4.1: xác nhận Lesson tồn tại, lưu file vật lý qua
// FileStorage, rồi ghi metadata vào bảng materials.
func (s *Service) Upload(ctx context.Context, lessonID uint, fileName string, content io.Reader) (*material.Material, error) {
	if _, err := s.lessons.FindByID(ctx, lessonID); err != nil {
		return nil, err
	}

	subdir := lessonSubdir(lessonID)
	path, err := s.storage.Save(ctx, subdir, fileName, content)
	if err != nil {
		return nil, err
	}

	m, err := material.NewMaterial(lessonID, fileName, path)
	if err != nil {
		return nil, err
	}
	if err := s.materials.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// ListByLesson hiện thực US4.2 (Học viên xem/tải tài liệu).
func (s *Service) ListByLesson(ctx context.Context, lessonID uint) ([]*material.Material, error) {
	return s.materials.ListByLesson(ctx, lessonID)
}

func lessonSubdir(lessonID uint) string {
	return "lesson-" + strconv.FormatUint(uint64(lessonID), 10)
}
