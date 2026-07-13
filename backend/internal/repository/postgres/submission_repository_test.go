package postgres_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/submission"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubmissionRepository(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, _ := user.NewUser("gv4@vlu.edu.vn", "GV", user.RoleTeacher)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	student, _ := user.NewUser("hv4@vlu.edu.vn", "Hoc Vien", user.RoleStudent)
	require.NoError(t, student.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, student))

	courseRepo := postgres.NewCourseRepository(pool)
	c, _ := course.NewCourse("Golang", "desc", teacher.ID())
	require.NoError(t, courseRepo.Create(ctx, c))

	chapterRepo := postgres.NewChapterRepository(pool)
	ch, _ := curriculum.NewChapter(c.ID(), "Chuong 1", 0)
	require.NoError(t, chapterRepo.Create(ctx, ch))

	lessonRepo := postgres.NewLessonRepository(pool)
	l, _ := curriculum.NewLesson(ch.ID(), "Bai 1", 0)
	require.NoError(t, lessonRepo.Create(ctx, l))

	assignmentRepo := postgres.NewAssignmentRepository(pool)
	quiz, _ := assignment.NewAssignment(l.ID(), "Trac nghiem", "", assignment.TypeQuiz, []assignment.Question{
		{Content: "1+1=?", Options: []string{"1", "2", "3"}, CorrectIndex: 1},
	}, nil, nil)
	require.NoError(t, assignmentRepo.Create(ctx, quiz))

	submissionRepo := postgres.NewSubmissionRepository(pool)

	_, err := submissionRepo.FindByAssignmentAndStudent(ctx, quiz.ID(), student.ID())
	assert.ErrorIs(t, err, submission.ErrNotFound)

	s, err := submission.NewSubmission(quiz.ID(), student.ID(), "", []int{1})
	require.NoError(t, err)
	require.NoError(t, submissionRepo.Create(ctx, s))
	assert.NotZero(t, s.ID())

	found, err := submissionRepo.FindByAssignmentAndStudent(ctx, quiz.ID(), student.ID())
	require.NoError(t, err)
	assert.Equal(t, []int{1}, found.Answers())

	list, err := submissionRepo.ListByAssignment(ctx, quiz.ID())
	require.NoError(t, err)
	require.Len(t, list, 1)

	byID, err := submissionRepo.FindByID(ctx, s.ID())
	require.NoError(t, err)
	assert.False(t, byID.IsGraded())

	require.NoError(t, byID.Grade(8.5, "Lam tot"))
	require.NoError(t, submissionRepo.Update(ctx, byID))

	regraded, err := submissionRepo.FindByID(ctx, s.ID())
	require.NoError(t, err)
	require.True(t, regraded.IsGraded())
	assert.Equal(t, 8.5, *regraded.Score())
	assert.Equal(t, "Lam tot", regraded.Feedback())
	assert.NotNil(t, regraded.GradedAt())

	_, err = submissionRepo.FindByID(ctx, 999999)
	assert.ErrorIs(t, err, submission.ErrNotFound)
}
