package postgres_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/enrollment"
	"educonnect-lms/backend/internal/domain/material"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrollmentRepository(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, _ := user.NewUser("gv@vlu.edu.vn", "GV", user.RoleTeacher)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	student, _ := user.NewUser("hv@vlu.edu.vn", "Hoc Vien A", user.RoleStudent)
	require.NoError(t, student.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, student))

	courseRepo := postgres.NewCourseRepository(pool)
	c, _ := course.NewCourse("Nhap mon Golang", "desc", teacher.ID())
	require.NoError(t, courseRepo.Create(ctx, c))

	enrollRepo := postgres.NewEnrollmentRepository(pool)

	isEnrolled, err := enrollRepo.IsEnrolled(ctx, student.ID(), c.ID())
	require.NoError(t, err)
	assert.False(t, isEnrolled)

	e, err := enrollment.NewEnrollment(student.ID(), c.ID())
	require.NoError(t, err)
	require.NoError(t, enrollRepo.Create(ctx, e))

	isEnrolled, err = enrollRepo.IsEnrolled(ctx, student.ID(), c.ID())
	require.NoError(t, err)
	assert.True(t, isEnrolled)

	students, err := enrollRepo.ListByCourse(ctx, c.ID())
	require.NoError(t, err)
	require.Len(t, students, 1)
	assert.Equal(t, student.ID(), students[0].StudentID())
}

func TestMaterialRepository(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, _ := user.NewUser("gv2@vlu.edu.vn", "GV", user.RoleTeacher)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	courseRepo := postgres.NewCourseRepository(pool)
	c, _ := course.NewCourse("Golang", "desc", teacher.ID())
	require.NoError(t, courseRepo.Create(ctx, c))

	chapterRepo := postgres.NewChapterRepository(pool)
	ch, _ := curriculum.NewChapter(c.ID(), "Chuong 1", 0)
	require.NoError(t, chapterRepo.Create(ctx, ch))

	lessonRepo := postgres.NewLessonRepository(pool)
	l, _ := curriculum.NewLesson(ch.ID(), "Bai 1", 0)
	require.NoError(t, lessonRepo.Create(ctx, l))

	materialRepo := postgres.NewMaterialRepository(pool)
	m, err := material.NewMaterial(l.ID(), "slide.pdf", "uploads/1/slide.pdf")
	require.NoError(t, err)
	require.NoError(t, materialRepo.Create(ctx, m))
	assert.NotZero(t, m.ID())

	list, err := materialRepo.ListByLesson(ctx, l.ID())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "slide.pdf", list[0].FileName())
}
