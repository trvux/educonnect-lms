// Package forum là service layer của US6.1: đăng và trả lời câu hỏi trong
// diễn đàn theo từng khóa học.
package forum

import (
	"context"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/forum"
)

// CourseGetter là tập con method của course.Repository, dùng để xác nhận
// khóa học tồn tại trước khi cho đăng bài.
type CourseGetter interface {
	FindByID(ctx context.Context, id uint) (*course.Course, error)
}

type Service struct {
	posts   forum.Repository
	courses CourseGetter
}

func NewService(posts forum.Repository, courses CourseGetter) *Service {
	return &Service{posts: posts, courses: courses}
}

// Post hiện thực US6.1: xác nhận khóa học tồn tại, và nếu là bài trả lời
// thì bài gốc phải cùng khóa học — trước khi ghi nhận bài đăng.
func (s *Service) Post(ctx context.Context, courseID, authorID uint, parentID *uint, content string) (*forum.Post, error) {
	if _, err := s.courses.FindByID(ctx, courseID); err != nil {
		return nil, err
	}

	if parentID != nil {
		parent, err := s.posts.FindByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent.CourseID() != courseID {
			return nil, forum.ErrParentCourseMismatch
		}
	}

	p, err := forum.NewPost(courseID, authorID, parentID, content)
	if err != nil {
		return nil, err
	}
	if err := s.posts.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListByCourse(ctx context.Context, courseID uint) ([]*forum.Post, error) {
	return s.posts.ListByCourse(ctx, courseID)
}
