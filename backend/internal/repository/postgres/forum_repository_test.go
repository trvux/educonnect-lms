package postgres_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/forum"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForumRepository(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, _ := user.NewUser("gv6@vlu.edu.vn", "GV", user.RoleTeacher)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	student, _ := user.NewUser("hv6@vlu.edu.vn", "Hoc Vien", user.RoleStudent)
	require.NoError(t, student.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, student))

	courseRepo := postgres.NewCourseRepository(pool)
	c, _ := course.NewCourse("Golang", "desc", teacher.ID())
	require.NoError(t, courseRepo.Create(ctx, c))

	forumRepo := postgres.NewForumRepository(pool)

	question, err := forum.NewPost(c.ID(), student.ID(), nil, "Bai tap nay lam sao vay thay?")
	require.NoError(t, err)
	require.NoError(t, forumRepo.Create(ctx, question))
	assert.NotZero(t, question.ID())

	replyParentID := question.ID()
	reply, err := forum.NewPost(c.ID(), teacher.ID(), &replyParentID, "Em xem lai muc 3 nhe")
	require.NoError(t, err)
	require.NoError(t, forumRepo.Create(ctx, reply))

	found, err := forumRepo.FindByID(ctx, question.ID())
	require.NoError(t, err)
	assert.Nil(t, found.ParentID())

	list, err := forumRepo.ListByCourse(ctx, c.ID())
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "Hoc Vien", list[0].AuthorName())
	require.NotNil(t, list[1].ParentID())
	assert.Equal(t, question.ID(), *list[1].ParentID())

	_, err = forumRepo.FindByID(ctx, 999999)
	assert.ErrorIs(t, err, forum.ErrNotFound)
}
