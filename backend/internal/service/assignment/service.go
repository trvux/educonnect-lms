// Package assignment là service layer của US5.1: Giảng viên tạo bài tập/
// trắc nghiệm gắn với 1 Lesson.
package assignment

import (
	"context"
	"time"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/curriculum"
)

type Service struct {
	assignments assignment.Repository
	lessons     curriculum.LessonRepository
}

func NewService(assignments assignment.Repository, lessons curriculum.LessonRepository) *Service {
	return &Service{assignments: assignments, lessons: lessons}
}

// Create hiện thực US5.1: xác nhận Lesson tồn tại rồi tạo bài tập/trắc
// nghiệm gắn với Lesson đó. timeLimitMinutes (US5.4) chỉ có ý nghĩa với bài
// trắc nghiệm — assignment.NewAssignment tự bỏ qua nếu kind là essay.
func (s *Service) Create(ctx context.Context, lessonID uint, title, description string, kind assignment.Type, questions []assignment.Question, dueAt *time.Time, timeLimitMinutes *int) (*assignment.Assignment, error) {
	if _, err := s.lessons.FindByID(ctx, lessonID); err != nil {
		return nil, err
	}

	a, err := assignment.NewAssignment(lessonID, title, description, kind, questions, dueAt, timeLimitMinutes)
	if err != nil {
		return nil, err
	}
	if err := s.assignments.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*assignment.Assignment, error) {
	return s.assignments.FindByID(ctx, id)
}

// ListByLesson hiện thực xem danh sách bài tập/trắc nghiệm của 1 Lesson
// (học viên và giảng viên đều xem được, chỉ khác ở việc ẩn/hiện đáp án
// đúng — xử lý ở tầng handler theo vai trò).
func (s *Service) ListByLesson(ctx context.Context, lessonID uint) ([]*assignment.Assignment, error) {
	return s.assignments.ListByLesson(ctx, lessonID)
}
