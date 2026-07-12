package curriculum_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/curriculum"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLesson(t *testing.T) {
	l, err := curriculum.NewLesson(1, "Bai 1: Gioi thieu Go", 0)
	require.NoError(t, err)
	assert.Equal(t, "Bai 1: Gioi thieu Go", l.Title())

	_, err = curriculum.NewLesson(0, "title", 0)
	assert.ErrorIs(t, err, curriculum.ErrInvalidChapterID)

	_, err = curriculum.NewLesson(1, "", 0)
	assert.ErrorIs(t, err, curriculum.ErrEmptyLessonTitle)
}
