// Package lessoncompletion là domain của US4.10: học viên đánh dấu đã học
// xong 1 bài học; bài học sau bị khóa cho tới khi bài trước hoàn thành.
//
// Tách riêng khỏi package progress (Epic 7, US7.1) vì đó là aggregate
// read-only tính % hoàn thành theo bài tập đã nộp cho dashboard, còn ở đây
// là 1 entity ghi nhận sự kiện "học viên X đã hoàn thành bài học Y" dùng để
// quyết định khóa/mở bài học tiếp theo — 2 khái niệm nghiệp vụ khác nhau.
package lessoncompletion

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("lesson completion: không tìm thấy")
	// ErrLessonLocked: học viên cố đánh dấu hoàn thành 1 bài học trong khi
	// bài học trước đó (theo thứ tự chương/bài trong khóa học) chưa hoàn
	// thành — chặn để giữ đúng tính tuần tự, kể cả khi UI (nút đã ẩn) bị
	// bỏ qua và gọi thẳng API.
	ErrLessonLocked = errors.New("lesson completion: bài học đang bị khóa, hoàn thành các bài học trước đó trước")
)

// LessonCompletion ghi nhận 1 học viên đã hoàn thành 1 bài học, tại 1 thời điểm.
type LessonCompletion struct {
	id          uint
	studentID   uint
	lessonID    uint
	completedAt time.Time
}

func New(studentID, lessonID uint) *LessonCompletion {
	return &LessonCompletion{studentID: studentID, lessonID: lessonID, completedAt: time.Now().UTC()}
}

func Rehydrate(id, studentID, lessonID uint, completedAt time.Time) *LessonCompletion {
	return &LessonCompletion{id: id, studentID: studentID, lessonID: lessonID, completedAt: completedAt}
}

func (c *LessonCompletion) ID() uint               { return c.id }
func (c *LessonCompletion) StudentID() uint        { return c.studentID }
func (c *LessonCompletion) LessonID() uint         { return c.lessonID }
func (c *LessonCompletion) CompletedAt() time.Time { return c.completedAt }

// LessonState là read model cho 1 bài học trong ngữ cảnh 1 học viên cụ thể:
// đã hoàn thành chưa, có đang bị khóa không (dùng để vẽ sidebar course
// player — US4.9/US4.10).
type LessonState struct {
	LessonID  uint
	Completed bool
	Locked    bool
}

// Repository là port mà service phụ thuộc, implement bởi internal/repository/postgres.
type Repository interface {
	// Create lưu 1 completion mới. Idempotent ở tầng DB (UNIQUE student_id,
	// lesson_id) — implementation coi vi phạm unique là thành công (đã
	// hoàn thành từ trước, ví dụ do race condition 2 request đồng thời).
	Create(ctx context.Context, c *LessonCompletion) error
	IsCompleted(ctx context.Context, studentID, lessonID uint) (bool, error)
	// ListCompletedByStudent trả về toàn bộ lesson_id mà student đã hoàn
	// thành (không giới hạn theo khóa học nào — service tự lọc theo lesson
	// nào liên quan khi dùng).
	ListCompletedByStudent(ctx context.Context, studentID uint) (map[uint]bool, error)
}
