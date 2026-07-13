package forum_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/forum"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPost_Question(t *testing.T) {
	p, err := forum.NewPost(1, 2, nil, "Bai tap nay lam sao vay thay?")
	require.NoError(t, err)
	assert.Equal(t, uint(1), p.CourseID())
	assert.Nil(t, p.ParentID())
}

func TestNewPost_Reply(t *testing.T) {
	parentID := uint(5)
	p, err := forum.NewPost(1, 2, &parentID, "Em xem lai muc 3 nhe")
	require.NoError(t, err)
	require.NotNil(t, p.ParentID())
	assert.Equal(t, uint(5), *p.ParentID())
}

func TestNewPost_Validation(t *testing.T) {
	_, err := forum.NewPost(0, 2, nil, "noi dung")
	assert.ErrorIs(t, err, forum.ErrInvalidCourseID)

	_, err = forum.NewPost(1, 0, nil, "noi dung")
	assert.ErrorIs(t, err, forum.ErrInvalidAuthorID)

	_, err = forum.NewPost(1, 2, nil, "")
	assert.ErrorIs(t, err, forum.ErrEmptyContent)
}
