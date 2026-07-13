package postgres_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChapterAndLessonRepository(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, err := user.NewUser("gv@vlu.edu.vn", "GV Huynh", user.RoleTeacher)
	require.NoError(t, err)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	courseRepo := postgres.NewCourseRepository(pool)
	c, err := course.NewCourse("Nhap mon Golang", "desc", teacher.ID())
	require.NoError(t, err)
	require.NoError(t, courseRepo.Create(ctx, c))

	chapterRepo := postgres.NewChapterRepository(pool)
	ch, err := curriculum.NewChapter(c.ID(), "Chuong 1: Nhap mon", 0)
	require.NoError(t, err)
	require.NoError(t, chapterRepo.Create(ctx, ch))
	assert.NotZero(t, ch.ID())

	count, err := chapterRepo.CountByCourse(ctx, c.ID())
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	lessonRepo := postgres.NewLessonRepository(pool)
	l, err := curriculum.NewLesson(ch.ID(), "Bai 1: Cai dat Go", 0)
	require.NoError(t, err)
	require.NoError(t, lessonRepo.Create(ctx, l))
	assert.NotZero(t, l.ID())

	lessons, err := lessonRepo.ListByChapter(ctx, ch.ID())
	require.NoError(t, err)
	require.Len(t, lessons, 1)
	assert.Equal(t, "Bai 1: Cai dat Go", lessons[0].Title())

	chapters, err := chapterRepo.ListByCourse(ctx, c.ID())
	require.NoError(t, err)
	require.Len(t, chapters, 1)
	assert.Equal(t, "Chuong 1: Nhap mon", chapters[0].Title())

	// US4.6 — không xóa được chương/bài học còn con bên trong (ràng buộc
	// khóa ngoại thật ở Postgres, dịch sang lỗi domain có ý nghĩa).
	err = chapterRepo.Delete(ctx, ch.ID())
	assert.ErrorIs(t, err, curriculum.ErrChapterNotEmpty, "chương còn bài học không được xóa")

	require.NoError(t, ch.Rename("Chuong 1 (da sua)"))
	require.NoError(t, chapterRepo.Update(ctx, ch))
	reloaded, err := chapterRepo.FindByID(ctx, ch.ID())
	require.NoError(t, err)
	assert.Equal(t, "Chuong 1 (da sua)", reloaded.Title())

	require.NoError(t, lessonRepo.Delete(ctx, l.ID()))
	_, err = lessonRepo.FindByID(ctx, l.ID())
	assert.ErrorIs(t, err, curriculum.ErrLessonNotFound, "bai hoc phai bi xoa that")

	require.NoError(t, chapterRepo.Delete(ctx, ch.ID()), "chuong rong (het bai hoc) phai xoa duoc")

	err = chapterRepo.Delete(ctx, ch.ID())
	assert.ErrorIs(t, err, curriculum.ErrChapterNotFound, "xoa lan 2 phai bao khong tim thay")
}
