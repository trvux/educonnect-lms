// Package quizattempt là service layer của US5.4: ghi nhận thời điểm học
// viên bắt đầu làm 1 bài trắc nghiệm có giới hạn thời gian.
package quizattempt

import (
	"context"
	"errors"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/quizattempt"
)

// AssignmentGetter là tập con method của *assignmentservice.Service.
type AssignmentGetter interface {
	Get(ctx context.Context, id uint) (*assignment.Assignment, error)
}

type Service struct {
	attempts    quizattempt.Repository
	assignments AssignmentGetter
}

func NewService(attempts quizattempt.Repository, assignments AssignmentGetter) *Service {
	return &Service{attempts: attempts, assignments: assignments}
}

// Start hiện thực US5.4: ghi nhận thời điểm học viên bắt đầu làm bài trắc
// nghiệm — idempotent, gọi lại (ví dụ refresh trang làm bài) trả về đúng
// attempt đã có từ trước, không "reset" đồng hồ đếm ngược. Không chặn nếu
// bài tập không phải trắc nghiệm hoặc không đặt giới hạn thời gian — cứ ghi
// nhận, vô hại; Frontend chỉ gọi endpoint này khi thật sự cần hiển thị đồng
// hồ đếm ngược (assignment.time_limit_minutes có giá trị).
func (s *Service) Start(ctx context.Context, assignmentID, studentID uint) (*quizattempt.QuizAttempt, error) {
	if _, err := s.assignments.Get(ctx, assignmentID); err != nil {
		return nil, err
	}

	existing, err := s.attempts.FindByAssignmentAndStudent(ctx, assignmentID, studentID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, quizattempt.ErrNotFound) {
		return nil, err
	}

	a := quizattempt.New(assignmentID, studentID)
	if err := s.attempts.Create(ctx, a); err != nil {
		return nil, err
	}
	// Phòng trường hợp Create bị race (2 request bắt đầu cùng lúc, unique
	// violation bị nuốt ở tầng repo) — đọc lại để chắc chắn lấy đúng attempt
	// gốc thay vì tin vào giá trị a (có thể không phải bản ghi thắng race).
	return s.attempts.FindByAssignmentAndStudent(ctx, assignmentID, studentID)
}
