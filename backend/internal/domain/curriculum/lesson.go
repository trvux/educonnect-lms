package curriculum

import (
	"context"
	"errors"
	"time"
)

var (
	ErrEmptyLessonTitle = errors.New("lesson: tiêu đề là bắt buộc")
	ErrInvalidChapterID = errors.New("lesson: chapter id là bắt buộc")
	ErrLessonNotFound   = errors.New("lesson: không tìm thấy")
	ErrLessonNotEmpty   = errors.New("lesson: còn tài liệu hoặc bài tập bên trong, xóa hết trước")
	// ErrInvalidLessonOrder (US4.7): tương tự ErrInvalidChapterOrder, nhưng
	// cho danh sách bài học trong 1 chương.
	ErrInvalidLessonOrder = errors.New("lesson: danh sách thứ tự không khớp với các bài học hiện có")
)

// Lesson là 1 bài học nằm trong 1 Chapter (US2.2). Tài liệu bài giảng
// (US4.1) sẽ gắn vào Lesson thông qua domain/material.
type Lesson struct {
	id        uint
	chapterID uint
	title     string
	position  int
	createdAt time.Time
	updatedAt time.Time
}

func NewLesson(chapterID uint, title string, position int) (*Lesson, error) {
	if chapterID == 0 {
		return nil, ErrInvalidChapterID
	}
	if title == "" {
		return nil, ErrEmptyLessonTitle
	}
	now := time.Now().UTC()
	return &Lesson{chapterID: chapterID, title: title, position: position, createdAt: now, updatedAt: now}, nil
}

func RehydrateLesson(id, chapterID uint, title string, position int, createdAt, updatedAt time.Time) *Lesson {
	return &Lesson{id: id, chapterID: chapterID, title: title, position: position, createdAt: createdAt, updatedAt: updatedAt}
}

func (l *Lesson) Rename(title string) error {
	if title == "" {
		return ErrEmptyLessonTitle
	}
	l.title = title
	l.updatedAt = time.Now().UTC()
	return nil
}

func (l *Lesson) Reorder(position int) {
	l.position = position
	l.updatedAt = time.Now().UTC()
}

func (l *Lesson) SetID(id uint) { l.id = id }

func (l *Lesson) ID() uint             { return l.id }
func (l *Lesson) ChapterID() uint      { return l.chapterID }
func (l *Lesson) Title() string        { return l.title }
func (l *Lesson) Position() int        { return l.position }
func (l *Lesson) CreatedAt() time.Time { return l.createdAt }
func (l *Lesson) UpdatedAt() time.Time { return l.updatedAt }

type LessonRepository interface {
	Create(ctx context.Context, l *Lesson) error
	FindByID(ctx context.Context, id uint) (*Lesson, error)
	ListByChapter(ctx context.Context, chapterID uint) ([]*Lesson, error)
	CountByChapter(ctx context.Context, chapterID uint) (int, error)
	Update(ctx context.Context, l *Lesson) error
	Delete(ctx context.Context, id uint) error
}
