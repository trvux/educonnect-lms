// Package notification là domain của US6.2: Giảng viên/Hệ thống gửi thông
// báo trong hệ thống đến học viên của khóa học.
package notification

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidRecipientID = errors.New("notification: recipient id là bắt buộc")
	ErrInvalidCourseID    = errors.New("notification: course id là bắt buộc")
	ErrEmptyTitle         = errors.New("notification: tiêu đề là bắt buộc")
	ErrNotFound           = errors.New("notification: không tìm thấy")
)

// Notification là 1 thông báo gửi tới 1 học viên, gắn với 1 khóa học
// (US6.2). Read đánh dấu học viên đã xem hay chưa — phục vụ badge số chưa
// đọc trên chuông thông báo của FE.
type Notification struct {
	id          uint
	recipientID uint
	courseID    uint
	title       string
	message     string
	read        bool
	createdAt   time.Time
}

func NewNotification(recipientID, courseID uint, title, message string) (*Notification, error) {
	if recipientID == 0 {
		return nil, ErrInvalidRecipientID
	}
	if courseID == 0 {
		return nil, ErrInvalidCourseID
	}
	if title == "" {
		return nil, ErrEmptyTitle
	}
	return &Notification{
		recipientID: recipientID,
		courseID:    courseID,
		title:       title,
		message:     message,
		read:        false,
		createdAt:   time.Now().UTC(),
	}, nil
}

func Rehydrate(id, recipientID, courseID uint, title, message string, read bool, createdAt time.Time) *Notification {
	return &Notification{
		id:          id,
		recipientID: recipientID,
		courseID:    courseID,
		title:       title,
		message:     message,
		read:        read,
		createdAt:   createdAt,
	}
}

func (n *Notification) SetID(id uint) { n.id = id }
func (n *Notification) MarkRead()     { n.read = true }

func (n *Notification) ID() uint             { return n.id }
func (n *Notification) RecipientID() uint    { return n.recipientID }
func (n *Notification) CourseID() uint       { return n.courseID }
func (n *Notification) Title() string        { return n.title }
func (n *Notification) Message() string      { return n.message }
func (n *Notification) Read() bool           { return n.read }
func (n *Notification) CreatedAt() time.Time { return n.createdAt }

type Repository interface {
	// CreateMany ghi 1 loạt thông báo trong 1 giao dịch (fan-out tới toàn
	// bộ học viên đã đăng ký khóa học, xem service/notification.SendToCourse).
	CreateMany(ctx context.Context, notifications []*Notification) error
	FindByID(ctx context.Context, id uint) (*Notification, error)
	ListByRecipient(ctx context.Context, recipientID uint) ([]*Notification, error)
	CountUnread(ctx context.Context, recipientID uint) (int, error)
	Update(ctx context.Context, n *Notification) error
}
