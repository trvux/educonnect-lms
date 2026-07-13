// Package assignment là domain của US5.1: Giảng viên tạo bài tập (tự luận,
// học viên nộp file) hoặc trắc nghiệm (nhiều câu hỏi trắc nghiệm) gắn với
// 1 Lesson cụ thể.
package assignment

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidLessonID    = errors.New("assignment: lesson id là bắt buộc")
	ErrEmptyTitle         = errors.New("assignment: tiêu đề là bắt buộc")
	ErrInvalidType        = errors.New("assignment: loại bài tập không hợp lệ (phải là essay hoặc quiz)")
	ErrQuizNeedsQuestions = errors.New("assignment: bài trắc nghiệm phải có ít nhất 1 câu hỏi")
	ErrInvalidQuestion    = errors.New("assignment: câu hỏi không hợp lệ (thiếu nội dung, dưới 2 lựa chọn, hoặc đáp án đúng nằm ngoài danh sách lựa chọn)")
	ErrNotFound           = errors.New("assignment: không tìm thấy")
	// ErrInvalidTimeLimit (US5.4): giới hạn thời gian (nếu có) phải là số
	// phút dương; chỉ áp dụng cho bài trắc nghiệm (essay không có khái niệm
	// "làm bài trong X phút" vì nộp file, không tính giờ).
	ErrInvalidTimeLimit = errors.New("assignment: giới hạn thời gian làm bài phải lớn hơn 0 phút")
)

// Type phân biệt bài tập tự luận (essay — học viên nộp file) và bài trắc
// nghiệm (quiz — chấm điểm tự động, xem US5.3).
type Type string

const (
	TypeEssay Type = "essay"
	TypeQuiz  Type = "quiz"
)

// Question là 1 câu hỏi trắc nghiệm gắn liền trong Assignment (child entity,
// không có repository riêng — lưu chung với Assignment dạng JSONB).
type Question struct {
	Content      string
	Options      []string
	CorrectIndex int
}

func (q Question) validate() error {
	if q.Content == "" {
		return ErrInvalidQuestion
	}
	if len(q.Options) < 2 {
		return ErrInvalidQuestion
	}
	if q.CorrectIndex < 0 || q.CorrectIndex >= len(q.Options) {
		return ErrInvalidQuestion
	}
	return nil
}

// Assignment đại diện cho 1 bài tập/trắc nghiệm gắn với 1 Lesson.
type Assignment struct {
	id               uint
	lessonID         uint
	title            string
	description      string
	kind             Type
	questions        []Question
	dueAt            *time.Time
	timeLimitMinutes *int
	createdAt        time.Time
}

func NewAssignment(lessonID uint, title, description string, kind Type, questions []Question, dueAt *time.Time, timeLimitMinutes *int) (*Assignment, error) {
	if lessonID == 0 {
		return nil, ErrInvalidLessonID
	}
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if kind != TypeEssay && kind != TypeQuiz {
		return nil, ErrInvalidType
	}
	if kind == TypeQuiz {
		if len(questions) == 0 {
			return nil, ErrQuizNeedsQuestions
		}
		for _, q := range questions {
			if err := q.validate(); err != nil {
				return nil, err
			}
		}
		if timeLimitMinutes != nil && *timeLimitMinutes <= 0 {
			return nil, ErrInvalidTimeLimit
		}
	} else {
		questions = nil
		timeLimitMinutes = nil // essay nộp file, không có khái niệm giới hạn thời gian
	}

	return &Assignment{
		lessonID:         lessonID,
		title:            title,
		description:      description,
		kind:             kind,
		questions:        questions,
		dueAt:            dueAt,
		timeLimitMinutes: timeLimitMinutes,
		createdAt:        time.Now().UTC(),
	}, nil
}

func Rehydrate(id, lessonID uint, title, description string, kind Type, questions []Question, dueAt *time.Time, timeLimitMinutes *int, createdAt time.Time) *Assignment {
	return &Assignment{
		id:               id,
		lessonID:         lessonID,
		title:            title,
		description:      description,
		kind:             kind,
		questions:        questions,
		dueAt:            dueAt,
		timeLimitMinutes: timeLimitMinutes,
		createdAt:        createdAt,
	}
}

func (a *Assignment) SetID(id uint) { a.id = id }

func (a *Assignment) ID() uint               { return a.id }
func (a *Assignment) LessonID() uint         { return a.lessonID }
func (a *Assignment) Title() string          { return a.title }
func (a *Assignment) Description() string    { return a.description }
func (a *Assignment) Kind() Type             { return a.kind }
func (a *Assignment) Questions() []Question  { return a.questions }
func (a *Assignment) DueAt() *time.Time      { return a.dueAt }
func (a *Assignment) TimeLimitMinutes() *int { return a.timeLimitMinutes }
func (a *Assignment) CreatedAt() time.Time   { return a.createdAt }

type Repository interface {
	Create(ctx context.Context, a *Assignment) error
	FindByID(ctx context.Context, id uint) (*Assignment, error)
	ListByLesson(ctx context.Context, lessonID uint) ([]*Assignment, error)
}
