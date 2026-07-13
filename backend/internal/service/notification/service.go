// Package notification là service layer của US6.2: Giảng viên/Hệ thống gửi
// thông báo trong hệ thống đến toàn bộ học viên đã đăng ký khóa học.
package notification

import (
	"context"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/enrollment"
	"educonnect-lms/backend/internal/domain/notification"
)

// CourseGetter là tập con method của course.Repository.
type CourseGetter interface {
	FindByID(ctx context.Context, id uint) (*course.Course, error)
}

// EnrollmentLister là tập con method của enrollment.Repository, dùng để
// xác định danh sách học viên nhận thông báo (fan-out).
type EnrollmentLister interface {
	ListByCourse(ctx context.Context, courseID uint) ([]*enrollment.Enrollment, error)
}

type Service struct {
	notifications notification.Repository
	enrollments   EnrollmentLister
	courses       CourseGetter
}

func NewService(notifications notification.Repository, enrollments EnrollmentLister, courses CourseGetter) *Service {
	return &Service{notifications: notifications, enrollments: enrollments, courses: courses}
}

// SendToCourse hiện thực US6.2: gửi 1 thông báo tới toàn bộ học viên đã
// đăng ký khóa học (fan-out — mỗi học viên nhận 1 bản ghi Notification
// riêng để tự đánh dấu đã đọc độc lập).
func (s *Service) SendToCourse(ctx context.Context, courseID uint, title, message string) ([]*notification.Notification, error) {
	if _, err := s.courses.FindByID(ctx, courseID); err != nil {
		return nil, err
	}

	enrollments, err := s.enrollments.ListByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	notifications := make([]*notification.Notification, 0, len(enrollments))
	for _, e := range enrollments {
		n, err := notification.NewNotification(e.StudentID(), courseID, title, message)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	if len(notifications) == 0 {
		return notifications, nil
	}
	if err := s.notifications.CreateMany(ctx, notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (s *Service) ListMine(ctx context.Context, recipientID uint) ([]*notification.Notification, error) {
	return s.notifications.ListByRecipient(ctx, recipientID)
}

func (s *Service) UnreadCount(ctx context.Context, recipientID uint) (int, error) {
	return s.notifications.CountUnread(ctx, recipientID)
}

// MarkRead chỉ cho phép chính chủ nhân thông báo đánh dấu đã đọc.
func (s *Service) MarkRead(ctx context.Context, id, recipientID uint) error {
	n, err := s.notifications.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if n.RecipientID() != recipientID {
		return notification.ErrNotFound
	}
	n.MarkRead()
	return s.notifications.Update(ctx, n)
}
