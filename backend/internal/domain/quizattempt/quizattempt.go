// Package quizattempt là domain của US5.4: ghi lại thời điểm học viên bắt
// đầu làm 1 bài trắc nghiệm có giới hạn thời gian, để tính đồng hồ đếm
// ngược ở phía server (không tin tưởng đồng hồ phía client, vì có thể bị
// giả mạo/reset khi refresh trang).
package quizattempt

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("quiz attempt: không tìm thấy")

// QuizAttempt ghi nhận 1 học viên đã bắt đầu làm 1 bài trắc nghiệm, tại 1
// thời điểm — không đổi qua các lần load lại trang (idempotent theo UNIQUE
// assignment_id, student_id), để đồng hồ đếm ngược không bị "reset" khi
// học viên refresh.
type QuizAttempt struct {
	id           uint
	assignmentID uint
	studentID    uint
	startedAt    time.Time
}

func New(assignmentID, studentID uint) *QuizAttempt {
	return &QuizAttempt{assignmentID: assignmentID, studentID: studentID, startedAt: time.Now().UTC()}
}

func Rehydrate(id, assignmentID, studentID uint, startedAt time.Time) *QuizAttempt {
	return &QuizAttempt{id: id, assignmentID: assignmentID, studentID: studentID, startedAt: startedAt}
}

func (a *QuizAttempt) SetID(id uint) { a.id = id }

func (a *QuizAttempt) ID() uint             { return a.id }
func (a *QuizAttempt) AssignmentID() uint   { return a.assignmentID }
func (a *QuizAttempt) StudentID() uint      { return a.studentID }
func (a *QuizAttempt) StartedAt() time.Time { return a.startedAt }

// Repository là port mà service phụ thuộc, implement bởi internal/repository/postgres.
type Repository interface {
	// Create lưu 1 attempt mới. Idempotent ở tầng DB (UNIQUE assignment_id,
	// student_id) — implementation coi vi phạm unique là thành công (đã có
	// attempt từ trước, ví dụ do race condition 2 tab cùng mở).
	Create(ctx context.Context, a *QuizAttempt) error
	FindByAssignmentAndStudent(ctx context.Context, assignmentID, studentID uint) (*QuizAttempt, error)
}
