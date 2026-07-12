// Package curriculum là service layer của US2.2: Giảng viên tạo/sắp xếp
// chương và bài học trong khóa học.
package curriculum

import (
	"context"

	"educonnect-lms/backend/internal/domain/curriculum"
)

type Service struct {
	chapters curriculum.ChapterRepository
	lessons  curriculum.LessonRepository
}

func NewService(chapters curriculum.ChapterRepository, lessons curriculum.LessonRepository) *Service {
	return &Service{chapters: chapters, lessons: lessons}
}

// CreateChapter tạo 1 chương mới, tự động xếp vào cuối danh sách chương
// hiện có của khóa học (US2.2).
func (s *Service) CreateChapter(ctx context.Context, courseID uint, title string) (*curriculum.Chapter, error) {
	count, err := s.chapters.CountByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	ch, err := curriculum.NewChapter(courseID, title, count)
	if err != nil {
		return nil, err
	}
	if err := s.chapters.Create(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *Service) ListChapters(ctx context.Context, courseID uint) ([]*curriculum.Chapter, error) {
	return s.chapters.ListByCourse(ctx, courseID)
}

// CreateLesson tạo 1 bài học mới trong 1 chương, tự xếp cuối danh sách bài
// học hiện có của chương đó (US2.2).
func (s *Service) CreateLesson(ctx context.Context, chapterID uint, title string) (*curriculum.Lesson, error) {
	// Đảm bảo chapter tồn tại trước khi tạo lesson con.
	if _, err := s.chapters.FindByID(ctx, chapterID); err != nil {
		return nil, err
	}
	count, err := s.lessons.CountByChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	l, err := curriculum.NewLesson(chapterID, title, count)
	if err != nil {
		return nil, err
	}
	if err := s.lessons.Create(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Service) ListLessons(ctx context.Context, chapterID uint) ([]*curriculum.Lesson, error) {
	return s.lessons.ListByChapter(ctx, chapterID)
}
