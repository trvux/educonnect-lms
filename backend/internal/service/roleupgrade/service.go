// Package roleupgrade là service layer của US1.7: Học viên gửi yêu cầu
// nâng cấp thành Giảng viên, Quản trị viên duyệt/từ chối.
package roleupgrade

import (
	"context"
	"errors"

	"educonnect-lms/backend/internal/domain/roleupgrade"
	"educonnect-lms/backend/internal/domain/user"
)

type Service struct {
	requests roleupgrade.Repository
	users    user.Repository
}

func NewService(requests roleupgrade.Repository, users user.Repository) *Service {
	return &Service{requests: requests, users: users}
}

// Create hiện thực US1.7: 1 học viên chỉ được có 1 yêu cầu đang chờ tại 1
// thời điểm.
func (s *Service) Create(ctx context.Context, userID uint, reason string) (*roleupgrade.Request, error) {
	_, err := s.requests.FindPendingByUser(ctx, userID)
	if err == nil {
		return nil, roleupgrade.ErrAlreadyPending
	}
	if !errors.Is(err, roleupgrade.ErrNotFound) {
		return nil, err
	}

	req, err := roleupgrade.NewRequest(userID, reason)
	if err != nil {
		return nil, err
	}
	if err := s.requests.Create(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Service) ListPending(ctx context.Context) ([]*roleupgrade.Request, error) {
	return s.requests.ListPending(ctx)
}

// Approve duyệt yêu cầu và đổi role user thành Teacher.
func (s *Service) Approve(ctx context.Context, requestID, adminID uint) (*roleupgrade.Request, error) {
	req, err := s.requests.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if err := req.Approve(adminID); err != nil {
		return nil, err
	}
	if err := s.requests.Update(ctx, req); err != nil {
		return nil, err
	}

	u, err := s.users.FindByID(ctx, req.UserID())
	if err != nil {
		return nil, err
	}
	if err := u.ChangeRole(user.RoleTeacher); err != nil {
		return nil, err
	}
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return req, nil
}

// Reject từ chối yêu cầu, giữ nguyên role Student.
func (s *Service) Reject(ctx context.Context, requestID, adminID uint) (*roleupgrade.Request, error) {
	req, err := s.requests.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if err := req.Reject(adminID); err != nil {
		return nil, err
	}
	if err := s.requests.Update(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}
