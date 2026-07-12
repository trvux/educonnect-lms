package course

import (
	"context"
	"errors"
	"time"
)

type Status string

const (
	StatusDraft    Status = "draft"          // US2.1: Giảng viên vừa tạo
	StatusPending  Status = "pending_review" // US2.1: Giảng viên gửi duyệt
	StatusApproved Status = "approved"       // US2.3: Quản trị viên duyệt -> công khai
)

var (
	ErrEmptyTitle       = errors.New("course: tiêu đề là bắt buộc")
	ErrInvalidTeacherID = errors.New("course: teacher id là bắt buộc")
	ErrNotPending       = errors.New("course: chỉ khóa học đang chờ duyệt mới được duyệt")
	ErrNotFound         = errors.New("course: không tìm thấy")
)

// Course là aggregate root cho Epic 2 (Quản lý khóa học).
type Course struct {
	id          uint
	title       string
	description string
	teacherID   uint
	status      Status
	createdAt   time.Time
	updatedAt   time.Time
}

// NewCourse tạo khóa học ở trạng thái Draft (US2.1). Luôn bắt đầu ở dạng
// chưa công khai; giảng viên phải chủ động SubmitForReview và quản trị viên
// phải Approve (US2.3) thì học viên mới tìm thấy được (US3.1 chỉ search course Approved).
func NewCourse(title, description string, teacherID uint) (*Course, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if teacherID == 0 {
		return nil, ErrInvalidTeacherID
	}
	now := time.Now().UTC()
	return &Course{
		title:       title,
		description: description,
		teacherID:   teacherID,
		status:      StatusDraft,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func Rehydrate(id uint, title, description string, teacherID uint, status Status, createdAt, updatedAt time.Time) *Course {
	return &Course{
		id:          id,
		title:       title,
		description: description,
		teacherID:   teacherID,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (c *Course) Rename(title, description string) error {
	if title == "" {
		return ErrEmptyTitle
	}
	c.title = title
	c.description = description
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Course) SubmitForReview() {
	c.status = StatusPending
	c.updatedAt = time.Now().UTC()
}

// Approve dùng cho US2.3 (Quản trị viên duyệt khóa học).
func (c *Course) Approve() error {
	if c.status != StatusPending {
		return ErrNotPending
	}
	c.status = StatusApproved
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Course) IsSearchable() bool { return c.status == StatusApproved }

func (c *Course) SetID(id uint) { c.id = id }

func (c *Course) ID() uint             { return c.id }
func (c *Course) Title() string        { return c.title }
func (c *Course) Description() string  { return c.description }
func (c *Course) TeacherID() uint      { return c.teacherID }
func (c *Course) Status() Status       { return c.status }
func (c *Course) CreatedAt() time.Time { return c.createdAt }
func (c *Course) UpdatedAt() time.Time { return c.updatedAt }

// Repository là port mà tầng service phụ thuộc vào (US2.1 tạo,
// US3.1 tìm kiếm); được implement bởi internal/repository/postgres.
type Repository interface {
	Create(ctx context.Context, c *Course) error
	FindByID(ctx context.Context, id uint) (*Course, error)
	Search(ctx context.Context, keyword string) ([]*Course, error)
	ListByTeacher(ctx context.Context, teacherID uint) ([]*Course, error)
	// ListByStatus dùng cho US2.3 (Admin xem hàng chờ duyệt PendingReview).
	ListByStatus(ctx context.Context, status Status) ([]*Course, error)
	Update(ctx context.Context, c *Course) error
}
