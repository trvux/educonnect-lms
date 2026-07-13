// Package roleupgrade là domain của US1.7: Học viên gửi yêu cầu nâng cấp
// thành Giảng viên, Quản trị viên duyệt hoặc từ chối — thay cho việc tự
// chọn vai trò công khai lúc đăng ký (không đúng chuẩn bảo mật LMS thật).
package roleupgrade

import (
	"context"
	"errors"
	"time"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

var (
	ErrEmptyReason    = errors.New("roleupgrade: lý do là bắt buộc")
	ErrAlreadyPending = errors.New("roleupgrade: bạn đã có 1 yêu cầu đang chờ duyệt")
	ErrNotPending     = errors.New("roleupgrade: yêu cầu không còn ở trạng thái chờ duyệt")
	ErrNotFound       = errors.New("roleupgrade: không tìm thấy")
)

// Request là 1 yêu cầu nâng cấp vai trò của 1 học viên.
type Request struct {
	id         uint
	userID     uint
	reason     string
	status     Status
	reviewedBy *uint
	createdAt  time.Time
	reviewedAt *time.Time
}

func NewRequest(userID uint, reason string) (*Request, error) {
	if reason == "" {
		return nil, ErrEmptyReason
	}
	return &Request{
		userID:    userID,
		reason:    reason,
		status:    StatusPending,
		createdAt: time.Now().UTC(),
	}, nil
}

func Rehydrate(id, userID uint, reason string, status Status, reviewedBy *uint, createdAt time.Time, reviewedAt *time.Time) *Request {
	return &Request{
		id:         id,
		userID:     userID,
		reason:     reason,
		status:     status,
		reviewedBy: reviewedBy,
		createdAt:  createdAt,
		reviewedAt: reviewedAt,
	}
}

func (r *Request) Approve(adminID uint) error { return r.review(StatusApproved, adminID) }
func (r *Request) Reject(adminID uint) error  { return r.review(StatusRejected, adminID) }

func (r *Request) review(next Status, adminID uint) error {
	if r.status != StatusPending {
		return ErrNotPending
	}
	r.status = next
	r.reviewedBy = &adminID
	now := time.Now().UTC()
	r.reviewedAt = &now
	return nil
}

func (r *Request) SetID(id uint) { r.id = id }

func (r *Request) ID() uint               { return r.id }
func (r *Request) UserID() uint           { return r.userID }
func (r *Request) Reason() string         { return r.reason }
func (r *Request) Status() Status         { return r.status }
func (r *Request) ReviewedBy() *uint      { return r.reviewedBy }
func (r *Request) CreatedAt() time.Time   { return r.createdAt }
func (r *Request) ReviewedAt() *time.Time { return r.reviewedAt }

type Repository interface {
	Create(ctx context.Context, r *Request) error
	FindByID(ctx context.Context, id uint) (*Request, error)
	// FindPendingByUser trả về ErrNotFound nếu user chưa có yêu cầu đang chờ
	// — dùng để chặn gửi trùng (US1.7).
	FindPendingByUser(ctx context.Context, userID uint) (*Request, error)
	ListPending(ctx context.Context) ([]*Request, error)
	Update(ctx context.Context, r *Request) error
}
