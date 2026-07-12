package course_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/course"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCourse(t *testing.T) {
	c, err := course.NewCourse("Nhap mon Golang", "Khoa hoc Golang co ban", 1)
	require.NoError(t, err)
	assert.Equal(t, course.StatusDraft, c.Status())
	assert.False(t, c.IsSearchable(), "draft course must not be searchable yet")

	_, err = course.NewCourse("", "desc", 1)
	assert.ErrorIs(t, err, course.ErrEmptyTitle)

	_, err = course.NewCourse("title", "desc", 0)
	assert.ErrorIs(t, err, course.ErrInvalidTeacherID)
}

func TestCourse_ApprovalFlow(t *testing.T) {
	c, err := course.NewCourse("Nhap mon Golang", "desc", 1)
	require.NoError(t, err)

	// Cannot approve directly from Draft.
	err = c.Approve()
	assert.ErrorIs(t, err, course.ErrNotPending)
	assert.False(t, c.IsSearchable())

	c.SubmitForReview()
	assert.Equal(t, course.StatusPending, c.Status())

	require.NoError(t, c.Approve())
	assert.Equal(t, course.StatusApproved, c.Status())
	assert.True(t, c.IsSearchable(), "approved course must become searchable (US3.1)")
}
