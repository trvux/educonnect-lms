// Package progress là service layer của US7.1: Học viên xem dashboard
// tiến độ học tập của mình.
package progress

import (
	"context"

	"educonnect-lms/backend/internal/domain/progress"
)

type Service struct {
	repo progress.Repository
}

func NewService(repo progress.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ForStudent(ctx context.Context, studentID uint) ([]progress.CourseProgress, error) {
	return s.repo.ForStudent(ctx, studentID)
}
