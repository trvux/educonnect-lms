// Package enrollment là domain của Epic 3: Học viên đăng ký khóa học
// (US3.2) và Giảng viên xem danh sách học viên trong lớp (US3.3).
package enrollment

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidStudentID = errors.New("enrollment: student id là bắt buộc")
	ErrInvalidCourseID  = errors.New("enrollment: course id là bắt buộc")
	ErrAlreadyEnrolled  = errors.New("enrollment: học viên đã đăng ký khóa học này rồi")
)

// Enrollment ghi nhận việc 1 học viên đã đăng ký 1 khóa học (US3.2).
type Enrollment struct {
	id         uint
	studentID  uint
	courseID   uint
	enrolledAt time.Time
}

func NewEnrollment(studentID, courseID uint) (*Enrollment, error) {
	if studentID == 0 {
		return nil, ErrInvalidStudentID
	}
	if courseID == 0 {
		return nil, ErrInvalidCourseID
	}
	return &Enrollment{studentID: studentID, courseID: courseID, enrolledAt: time.Now().UTC()}, nil
}

func Rehydrate(id, studentID, courseID uint, enrolledAt time.Time) *Enrollment {
	return &Enrollment{id: id, studentID: studentID, courseID: courseID, enrolledAt: enrolledAt}
}

func (e *Enrollment) SetID(id uint) { e.id = id }

func (e *Enrollment) ID() uint              { return e.id }
func (e *Enrollment) StudentID() uint       { return e.studentID }
func (e *Enrollment) CourseID() uint        { return e.courseID }
func (e *Enrollment) EnrolledAt() time.Time { return e.enrolledAt }

// Repository là port mà service phụ thuộc vào, implement bởi
// internal/repository/postgres. IsEnrolled dùng để chặn đăng ký trùng
// (US3.2) trước khi tạo bản ghi mới.
type Repository interface {
	Create(ctx context.Context, e *Enrollment) error
	IsEnrolled(ctx context.Context, studentID, courseID uint) (bool, error)
	ListByCourse(ctx context.Context, courseID uint) ([]*Enrollment, error)
	ListByStudent(ctx context.Context, studentID uint) ([]*Enrollment, error)
}
