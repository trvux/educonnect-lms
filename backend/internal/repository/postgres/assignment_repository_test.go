package postgres_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssignmentRepository(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, _ := user.NewUser("gv3@vlu.edu.vn", "GV", user.RoleTeacher)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	courseRepo := postgres.NewCourseRepository(pool)
	c, _ := course.NewCourse("Golang nang cao", "desc", teacher.ID())
	require.NoError(t, courseRepo.Create(ctx, c))

	chapterRepo := postgres.NewChapterRepository(pool)
	ch, _ := curriculum.NewChapter(c.ID(), "Chuong 1", 0)
	require.NoError(t, chapterRepo.Create(ctx, ch))

	lessonRepo := postgres.NewLessonRepository(pool)
	l, _ := curriculum.NewLesson(ch.ID(), "Bai 1", 0)
	require.NoError(t, lessonRepo.Create(ctx, l))

	assignmentRepo := postgres.NewAssignmentRepository(pool)

	essay, err := assignment.NewAssignment(l.ID(), "Bai tap tu luan", "Nop file .go", assignment.TypeEssay, nil, nil, nil)
	require.NoError(t, err)
	require.NoError(t, assignmentRepo.Create(ctx, essay))
	assert.NotZero(t, essay.ID())

	quiz, err := assignment.NewAssignment(l.ID(), "Trac nghiem chuong 1", "", assignment.TypeQuiz, []assignment.Question{
		{Content: "1 + 1 = ?", Options: []string{"1", "2", "3"}, CorrectIndex: 1},
	}, nil, nil)
	require.NoError(t, err)
	require.NoError(t, assignmentRepo.Create(ctx, quiz))

	found, err := assignmentRepo.FindByID(ctx, quiz.ID())
	require.NoError(t, err)
	assert.Equal(t, "Trac nghiem chuong 1", found.Title())
	require.Len(t, found.Questions(), 1)
	assert.Equal(t, "1 + 1 = ?", found.Questions()[0].Content)
	assert.Equal(t, 1, found.Questions()[0].CorrectIndex)

	list, err := assignmentRepo.ListByLesson(ctx, l.ID())
	require.NoError(t, err)
	require.Len(t, list, 2)

	_, err = assignmentRepo.FindByID(ctx, 999999)
	assert.ErrorIs(t, err, assignment.ErrNotFound)
}
