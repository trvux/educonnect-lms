// Package curriculum chứa domain của Epic 2 phần cấu trúc nội dung khóa học:
// Chapter (chương) chứa nhiều Lesson (bài học), thuộc về 1 Course (US2.2).
package curriculum

import (
	"context"
	"errors"
	"time"
)

var (
	ErrEmptyChapterTitle = errors.New("chapter: tiêu đề là bắt buộc")
	ErrInvalidCourseID   = errors.New("chapter: course id là bắt buộc")
	ErrChapterNotFound   = errors.New("chapter: không tìm thấy")
)

// Chapter là 1 chương trong khóa học, có thứ tự hiển thị (Position) để
// giảng viên sắp xếp trên UI (kéo-thả).
type Chapter struct {
	id        uint
	courseID  uint
	title     string
	position  int
	createdAt time.Time
	updatedAt time.Time
}

// NewChapter tạo 1 chương mới (US2.2). Position do caller (service) tính
// dựa trên số chương hiện có của khóa học, để chương mới luôn nằm cuối danh sách.
func NewChapter(courseID uint, title string, position int) (*Chapter, error) {
	if courseID == 0 {
		return nil, ErrInvalidCourseID
	}
	if title == "" {
		return nil, ErrEmptyChapterTitle
	}
	now := time.Now().UTC()
	return &Chapter{courseID: courseID, title: title, position: position, createdAt: now, updatedAt: now}, nil
}

func RehydrateChapter(id, courseID uint, title string, position int, createdAt, updatedAt time.Time) *Chapter {
	return &Chapter{id: id, courseID: courseID, title: title, position: position, createdAt: createdAt, updatedAt: updatedAt}
}

func (c *Chapter) Rename(title string) error {
	if title == "" {
		return ErrEmptyChapterTitle
	}
	c.title = title
	c.updatedAt = time.Now().UTC()
	return nil
}

// Reorder được dùng khi giảng viên kéo-thả sắp xếp lại thứ tự chương (US2.2).
func (c *Chapter) Reorder(position int) {
	c.position = position
	c.updatedAt = time.Now().UTC()
}

func (c *Chapter) SetID(id uint) { c.id = id }

func (c *Chapter) ID() uint             { return c.id }
func (c *Chapter) CourseID() uint       { return c.courseID }
func (c *Chapter) Title() string        { return c.title }
func (c *Chapter) Position() int        { return c.position }
func (c *Chapter) CreatedAt() time.Time { return c.createdAt }
func (c *Chapter) UpdatedAt() time.Time { return c.updatedAt }

// ChapterRepository là port mà service phụ thuộc, implement bởi
// internal/repository/postgres.
type ChapterRepository interface {
	Create(ctx context.Context, c *Chapter) error
	FindByID(ctx context.Context, id uint) (*Chapter, error)
	ListByCourse(ctx context.Context, courseID uint) ([]*Chapter, error)
	CountByCourse(ctx context.Context, courseID uint) (int, error)
	Update(ctx context.Context, c *Chapter) error
}
