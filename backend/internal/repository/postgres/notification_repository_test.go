package postgres_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/notification"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationRepository(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, _ := user.NewUser("gv7@vlu.edu.vn", "GV", user.RoleTeacher)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	student1, _ := user.NewUser("hv7a@vlu.edu.vn", "Hoc Vien A", user.RoleStudent)
	require.NoError(t, student1.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, student1))

	student2, _ := user.NewUser("hv7b@vlu.edu.vn", "Hoc Vien B", user.RoleStudent)
	require.NoError(t, student2.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, student2))

	courseRepo := postgres.NewCourseRepository(pool)
	c, _ := course.NewCourse("Golang", "desc", teacher.ID())
	require.NoError(t, courseRepo.Create(ctx, c))

	notificationRepo := postgres.NewNotificationRepository(pool)

	n1, _ := notification.NewNotification(student1.ID(), c.ID(), "Bai tap moi", "Co bai tap moi")
	n2, _ := notification.NewNotification(student2.ID(), c.ID(), "Bai tap moi", "Co bai tap moi")
	require.NoError(t, notificationRepo.CreateMany(ctx, []*notification.Notification{n1, n2}))
	assert.NotZero(t, n1.ID())
	assert.NotZero(t, n2.ID())

	list, err := notificationRepo.ListByRecipient(ctx, student1.ID())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.False(t, list[0].Read())

	unread, err := notificationRepo.CountUnread(ctx, student1.ID())
	require.NoError(t, err)
	assert.Equal(t, 1, unread)

	found, err := notificationRepo.FindByID(ctx, n1.ID())
	require.NoError(t, err)
	found.MarkRead()
	require.NoError(t, notificationRepo.Update(ctx, found))

	unread, err = notificationRepo.CountUnread(ctx, student1.ID())
	require.NoError(t, err)
	assert.Equal(t, 0, unread)

	_, err = notificationRepo.FindByID(ctx, 999999)
	assert.ErrorIs(t, err, notification.ErrNotFound)
}
