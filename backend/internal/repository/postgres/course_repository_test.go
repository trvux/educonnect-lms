package postgres_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCourseRepository_CreateAndSearch(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, err := user.NewUser("gv@vlu.edu.vn", "GV Huynh", user.RoleTeacher)
	require.NoError(t, err)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	courseRepo := postgres.NewCourseRepository(pool)

	c, err := course.NewCourse("Nhap mon Golang", "Khoa hoc Golang co ban", teacher.ID())
	require.NoError(t, err)
	require.NoError(t, courseRepo.Create(ctx, c))
	assert.NotZero(t, c.ID())

	// Draft courses must not show up in search (US3.1 only shows Approved).
	results, err := courseRepo.Search(ctx, "Golang")
	require.NoError(t, err)
	assert.Empty(t, results)

	c.SubmitForReview()
	require.NoError(t, c.Approve())
	require.NoError(t, courseRepo.Update(ctx, c))

	results, err = courseRepo.Search(ctx, "golang") // case-insensitive (ILIKE)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Nhap mon Golang", results[0].Title())
}

func TestCourseRepository_ListByTeacher(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, err := user.NewUser("gv2@vlu.edu.vn", "GV Tran", user.RoleTeacher)
	require.NoError(t, err)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	courseRepo := postgres.NewCourseRepository(pool)
	c1, _ := course.NewCourse("Course A", "", teacher.ID())
	c2, _ := course.NewCourse("Course B", "", teacher.ID())
	require.NoError(t, courseRepo.Create(ctx, c1))
	require.NoError(t, courseRepo.Create(ctx, c2))

	list, err := courseRepo.ListByTeacher(ctx, teacher.ID())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}
