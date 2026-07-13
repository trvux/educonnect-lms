// Package submission là service layer của US5.2: Học viên nộp bài tập tự
// luận hoặc làm bài trắc nghiệm trước hạn.
package submission

import (
	"context"
	"errors"
	"time"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/submission"
)

// AssignmentGetter là tập con method của *assignmentservice.Service mà
// service này cần để kiểm tra hạn nộp/số câu hỏi.
type AssignmentGetter interface {
	Get(ctx context.Context, id uint) (*assignment.Assignment, error)
}

type Service struct {
	submissions submission.Repository
	assignments AssignmentGetter
}

func NewService(submissions submission.Repository, assignments AssignmentGetter) *Service {
	return &Service{submissions: submissions, assignments: assignments}
}

// Submit hiện thực US5.2: xác nhận Assignment tồn tại, chưa quá hạn nộp,
// học viên chưa nộp bài này trước đó, và (với bài trắc nghiệm) số đáp án
// khớp số câu hỏi — trước khi ghi nhận bài làm.
func (s *Service) Submit(ctx context.Context, assignmentID, studentID uint, content string, answers []int) (*submission.Submission, error) {
	a, err := s.assignments.Get(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	if a.DueAt() != nil && time.Now().UTC().After(*a.DueAt()) {
		return nil, submission.ErrPastDue
	}

	if a.Kind() == assignment.TypeQuiz && len(answers) != len(a.Questions()) {
		return nil, submission.ErrAnswerCountMismatch
	}

	_, err = s.submissions.FindByAssignmentAndStudent(ctx, assignmentID, studentID)
	if err == nil {
		return nil, submission.ErrAlreadySubmitted
	}
	if !errors.Is(err, submission.ErrNotFound) {
		return nil, err
	}

	sub, err := submission.NewSubmission(assignmentID, studentID, content, answers)
	if err != nil {
		return nil, err
	}
	if err := s.submissions.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}
