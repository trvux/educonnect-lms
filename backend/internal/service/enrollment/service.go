// Package enrollment là service layer của US3.2 (Học viên đăng ký khóa
// học) và US3.3 (Giảng viên xem danh sách học viên trong lớp).
package enrollment

import (
	"context"
	"errors"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/enrollment"
	"educonnect-lms/backend/internal/domain/user"
)

var ErrCourseNotApproved = errors.New("enrollment: chỉ khóa học đã duyệt (approved) mới cho đăng ký")

type Service struct {
	enrollments enrollment.Repository
	courses     course.Repository
	users       user.Repository
}

func NewService(enrollments enrollment.Repository, courses course.Repository, users user.Repository) *Service {
	return &Service{enrollments: enrollments, courses: courses, users: users}
}

// Enroll hiện thực US3.2. Chỉ khóa học đã Approved mới cho đăng ký, và
// không cho đăng ký trùng (IsEnrolled check trước khi Create).
func (s *Service) Enroll(ctx context.Context, studentID, courseID uint) (*enrollment.Enrollment, error) {
	c, err := s.courses.FindByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if !c.IsSearchable() { // IsSearchable == đã Approved, xem domain/course
		return nil, ErrCourseNotApproved
	}

	already, err := s.enrollments.IsEnrolled(ctx, studentID, courseID)
	if err != nil {
		return nil, err
	}
	if already {
		return nil, enrollment.ErrAlreadyEnrolled
	}

	e, err := enrollment.NewEnrollment(studentID, courseID)
	if err != nil {
		return nil, err
	}
	if err := s.enrollments.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

// EnrolledStudent gộp thông tin Enrollment + User để trả về cho UI (US3.3),
// tránh handler phải tự join dữ liệu.
type EnrolledStudent struct {
	StudentID uint
	FullName  string
	Email     string
}

// ListStudents hiện thực US3.3: chỉ giảng viên sở hữu khóa học mới được
// xem danh sách học viên trong lớp.
func (s *Service) ListStudents(ctx context.Context, courseID, requestingTeacherID uint) ([]EnrolledStudent, error) {
	c, err := s.courses.FindByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if c.TeacherID() != requestingTeacherID {
		return nil, course.ErrNotFound // không tiết lộ khóa học của giáo viên khác
	}

	enrollments, err := s.enrollments.ListByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	result := make([]EnrolledStudent, 0, len(enrollments))
	for _, e := range enrollments {
		u, err := s.users.FindByID(ctx, e.StudentID())
		if err != nil {
			continue // học viên đã bị xoá/không tìm thấy, bỏ qua thay vì fail cả danh sách
		}
		result = append(result, EnrolledStudent{StudentID: u.ID(), FullName: u.FullName(), Email: u.Email()})
	}
	return result, nil
}
