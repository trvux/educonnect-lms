package course

import (
	"context"

	"educonnect-lms/backend/internal/domain/course"
)

type Service struct {
	courses course.Repository
}

func NewService(courses course.Repository) *Service {
	return &Service{courses: courses}
}

type CreateInput struct {
	Title       string
	Description string
	TeacherID   uint
}

// Create implements US2.1 (Giảng viên tạo khóa học). New courses always
// start as Draft — SubmitForReview/Approve are separate operations.
func (s *Service) Create(ctx context.Context, in CreateInput) (*course.Course, error) {
	c, err := course.NewCourse(in.Title, in.Description, in.TeacherID)
	if err != nil {
		return nil, err
	}
	if err := s.courses.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Search implements US3.1 (Học viên tìm kiếm khóa học) — only Approved
// courses are returned, enforced by the repository query.
func (s *Service) Search(ctx context.Context, keyword string) ([]*course.Course, error) {
	return s.courses.Search(ctx, keyword)
}

func (s *Service) ListByTeacher(ctx context.Context, teacherID uint) ([]*course.Course, error) {
	return s.courses.ListByTeacher(ctx, teacherID)
}

// SubmitForReview lets the owning teacher move a Draft course to
// PendingReview so it appears in the admin approval queue (still part of
// US2.1's authoring flow, precedes US2.3).
func (s *Service) SubmitForReview(ctx context.Context, courseID, teacherID uint) (*course.Course, error) {
	c, err := s.courses.FindByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if c.TeacherID() != teacherID {
		return nil, course.ErrNotFound // do not leak existence of other teachers' courses
	}
	c.SubmitForReview()
	if err := s.courses.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Approve implements US2.3 (Quản trị viên duyệt khóa học). It only succeeds
// for courses already in PendingReview — enforced by course.Course.Approve.
func (s *Service) Approve(ctx context.Context, courseID uint) (*course.Course, error) {
	c, err := s.courses.FindByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if err := c.Approve(); err != nil {
		return nil, err
	}
	if err := s.courses.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}
