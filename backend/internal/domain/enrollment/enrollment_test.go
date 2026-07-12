package enrollment_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/enrollment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnrollment(t *testing.T) {
	e, err := enrollment.NewEnrollment(10, 20)
	require.NoError(t, err)
	assert.Equal(t, uint(10), e.StudentID())
	assert.Equal(t, uint(20), e.CourseID())
	assert.False(t, e.EnrolledAt().IsZero())

	_, err = enrollment.NewEnrollment(0, 20)
	assert.ErrorIs(t, err, enrollment.ErrInvalidStudentID)

	_, err = enrollment.NewEnrollment(10, 0)
	assert.ErrorIs(t, err, enrollment.ErrInvalidCourseID)
}
