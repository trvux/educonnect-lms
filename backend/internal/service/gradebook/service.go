// Package gradebook là service layer của phần "hệ thống tự tổng hợp bảng
// điểm" trong US5.3.
package gradebook

import (
	"context"

	"educonnect-lms/backend/internal/domain/gradebook"
)

type Service struct {
	entries gradebook.Repository
}

func NewService(entries gradebook.Repository) *Service {
	return &Service{entries: entries}
}

func (s *Service) ForCourse(ctx context.Context, courseID uint) ([]gradebook.Entry, error) {
	return s.entries.ForCourse(ctx, courseID)
}
