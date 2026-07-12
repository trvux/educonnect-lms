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

// Create hiện thực US2.1 (Giảng viên tạo khóa học). Khóa học mới luôn
// bắt đầu ở Draft — SubmitForReview/Approve là 2 thao tác riêng biệt.
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

// Search hiện thực US3.1 (Học viên tìm kiếm khóa học) — chỉ khóa học
// đã Approved mới được trả về, thực thi bởi query trong repository.
func (s *Service) Search(ctx context.Context, keyword string) ([]*course.Course, error) {
	return s.courses.Search(ctx, keyword)
}

// Get trả về 1 khóa học theo ID (dùng cho trang chi tiết khóa học).
func (s *Service) Get(ctx context.Context, id uint) (*course.Course, error) {
	return s.courses.FindByID(ctx, id)
}

// ListPending hiện thực hàng chờ duyệt của US2.3 (Admin xem danh sách
// khóa học đang PendingReview trước khi Approve).
func (s *Service) ListPending(ctx context.Context) ([]*course.Course, error) {
	return s.courses.ListByStatus(ctx, course.StatusPending)
}

func (s *Service) ListByTeacher(ctx context.Context, teacherID uint) ([]*course.Course, error) {
	return s.courses.ListByTeacher(ctx, teacherID)
}

// SubmitForReview cho phép giảng viên sở hữu khóa học chuyển nó từ Draft
// sang PendingReview để xuất hiện trong hàng chờ duyệt của admin (vẫn thuộc
// luồng soạn thảo của US2.1, diễn ra trước US2.3).
func (s *Service) SubmitForReview(ctx context.Context, courseID, teacherID uint) (*course.Course, error) {
	c, err := s.courses.FindByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if c.TeacherID() != teacherID {
		return nil, course.ErrNotFound // không tiết lộ sự tồn tại của khóa học người khác
	}
	c.SubmitForReview()
	if err := s.courses.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Approve hiện thực US2.3 (Quản trị viên duyệt khóa học). Chỉ thành công
// với khóa học đang ở PendingReview — được đảm bảo bởi course.Course.Approve.
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
