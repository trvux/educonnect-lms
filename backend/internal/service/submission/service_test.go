package submission_test

import (
	"context"
	"testing"
	"time"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/submission"
	submissionservice "educonnect-lms/backend/internal/service/submission"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAssignmentGetter struct {
	items map[uint]*assignment.Assignment
}

func (r *fakeAssignmentGetter) Get(_ context.Context, id uint) (*assignment.Assignment, error) {
	a, ok := r.items[id]
	if !ok {
		return nil, assignment.ErrNotFound
	}
	return a, nil
}

type fakeSubmissionRepo struct {
	items  []*submission.Submission
	nextID uint
}

func (r *fakeSubmissionRepo) Create(_ context.Context, s *submission.Submission) error {
	r.nextID++
	s.SetID(r.nextID)
	r.items = append(r.items, s)
	return nil
}
func (r *fakeSubmissionRepo) FindByAssignmentAndStudent(_ context.Context, assignmentID, studentID uint) (*submission.Submission, error) {
	for _, s := range r.items {
		if s.AssignmentID() == assignmentID && s.StudentID() == studentID {
			return s, nil
		}
	}
	return nil, submission.ErrNotFound
}
func (r *fakeSubmissionRepo) ListByAssignment(_ context.Context, assignmentID uint) ([]*submission.Submission, error) {
	var out []*submission.Submission
	for _, s := range r.items {
		if s.AssignmentID() == assignmentID {
			out = append(out, s)
		}
	}
	return out, nil
}

func newQuiz(id uint, dueAt *time.Time) *assignment.Assignment {
	a, _ := assignment.NewAssignment(1, "Trac nghiem", "", assignment.TypeQuiz, []assignment.Question{
		{Content: "1+1=?", Options: []string{"1", "2", "3"}, CorrectIndex: 1},
	}, dueAt)
	a.SetID(id)
	return a
}

func TestService_Submit(t *testing.T) {
	assignments := &fakeAssignmentGetter{items: map[uint]*assignment.Assignment{1: newQuiz(1, nil)}}
	submissions := &fakeSubmissionRepo{}
	svc := submissionservice.NewService(submissions, assignments)

	s, err := svc.Submit(context.Background(), 1, 2, "", []int{1})
	require.NoError(t, err)
	assert.NotZero(t, s.ID())
}

func TestService_Submit_AssignmentNotFound(t *testing.T) {
	svc := submissionservice.NewService(&fakeSubmissionRepo{}, &fakeAssignmentGetter{items: map[uint]*assignment.Assignment{}})

	_, err := svc.Submit(context.Background(), 999, 2, "", []int{1})
	assert.ErrorIs(t, err, assignment.ErrNotFound)
}

func TestService_Submit_PastDue(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour)
	assignments := &fakeAssignmentGetter{items: map[uint]*assignment.Assignment{1: newQuiz(1, &past)}}
	svc := submissionservice.NewService(&fakeSubmissionRepo{}, assignments)

	_, err := svc.Submit(context.Background(), 1, 2, "", []int{1})
	assert.ErrorIs(t, err, submission.ErrPastDue)
}

func TestService_Submit_AnswerCountMismatch(t *testing.T) {
	assignments := &fakeAssignmentGetter{items: map[uint]*assignment.Assignment{1: newQuiz(1, nil)}}
	svc := submissionservice.NewService(&fakeSubmissionRepo{}, assignments)

	_, err := svc.Submit(context.Background(), 1, 2, "", []int{1, 0})
	assert.ErrorIs(t, err, submission.ErrAnswerCountMismatch)
}

func TestService_Submit_AlreadySubmitted(t *testing.T) {
	assignments := &fakeAssignmentGetter{items: map[uint]*assignment.Assignment{1: newQuiz(1, nil)}}
	submissions := &fakeSubmissionRepo{}
	svc := submissionservice.NewService(submissions, assignments)

	_, err := svc.Submit(context.Background(), 1, 2, "", []int{1})
	require.NoError(t, err)

	_, err = svc.Submit(context.Background(), 1, 2, "", []int{0})
	assert.ErrorIs(t, err, submission.ErrAlreadySubmitted)
}
