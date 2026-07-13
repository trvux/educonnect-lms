package submission_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/submission"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubmission_Essay(t *testing.T) {
	s, err := submission.NewSubmission(1, 2, "bai lam cua toi", nil)
	require.NoError(t, err)
	assert.Equal(t, "bai lam cua toi", s.Content())
	assert.Empty(t, s.Answers())
}

func TestNewSubmission_Quiz(t *testing.T) {
	s, err := submission.NewSubmission(1, 2, "", []int{1, 0, 2})
	require.NoError(t, err)
	assert.Equal(t, []int{1, 0, 2}, s.Answers())
}

func TestNewSubmission_Validation(t *testing.T) {
	_, err := submission.NewSubmission(0, 2, "bai lam", nil)
	assert.ErrorIs(t, err, submission.ErrInvalidAssignmentID)

	_, err = submission.NewSubmission(1, 0, "bai lam", nil)
	assert.ErrorIs(t, err, submission.ErrInvalidStudentID)

	_, err = submission.NewSubmission(1, 2, "", nil)
	assert.ErrorIs(t, err, submission.ErrEmptySubmission)
}
