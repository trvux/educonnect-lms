package curriculum_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/curriculum"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChapter(t *testing.T) {
	c, err := curriculum.NewChapter(1, "Chuong 1: Nhap mon", 0)
	require.NoError(t, err)
	assert.Equal(t, "Chuong 1: Nhap mon", c.Title())
	assert.Equal(t, 0, c.Position())

	_, err = curriculum.NewChapter(0, "title", 0)
	assert.ErrorIs(t, err, curriculum.ErrInvalidCourseID)

	_, err = curriculum.NewChapter(1, "", 0)
	assert.ErrorIs(t, err, curriculum.ErrEmptyChapterTitle)
}

func TestChapter_ReorderAndRename(t *testing.T) {
	c, err := curriculum.NewChapter(1, "Chuong 1", 0)
	require.NoError(t, err)

	c.Reorder(2)
	assert.Equal(t, 2, c.Position())

	require.NoError(t, c.Rename("Chuong 1 (sua)"))
	assert.Equal(t, "Chuong 1 (sua)", c.Title())

	assert.ErrorIs(t, c.Rename(""), curriculum.ErrEmptyChapterTitle)
}
