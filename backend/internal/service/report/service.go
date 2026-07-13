// Package report là service layer của US7.2: Giảng viên/Quản trị viên xem
// báo cáo thống kê học viên/khóa học.
package report

import (
	"context"

	"educonnect-lms/backend/internal/domain/report"
)

type Service struct {
	repo report.Repository
}

func NewService(repo report.Repository) *Service {
	return &Service{repo: repo}
}

// ForTeacher trả về báo cáo mọi khóa học do giảng viên đó sở hữu.
func (s *Service) ForTeacher(ctx context.Context, teacherID uint) ([]report.CourseStats, error) {
	return s.repo.ForTeacher(ctx, teacherID)
}

// All trả về báo cáo mọi khóa học trong hệ thống (quản trị viên).
func (s *Service) All(ctx context.Context) ([]report.CourseStats, error) {
	return s.repo.All(ctx)
}
