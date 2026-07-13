package postgres_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/enrollment"
	"educonnect-lms/backend/internal/domain/submission"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGradebookRepository_ForCourse(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, _ := user.NewUser("gv5@vlu.edu.vn", "GV", user.RoleTeacher)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	student, _ := user.NewUser("hv5@vlu.edu.vn", "Hoc Vien Nam", user.RoleStudent)
	require.NoError(t, student.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, student))

	courseRepo := postgres.NewCourseRepository(pool)
	c, _ := course.NewCourse("Golang", "desc", teacher.ID())
	require.NoError(t, courseRepo.Create(ctx, c))

	enrollRepo := postgres.NewEnrollmentRepository(pool)
	e, _ := enrollment.NewEnrollment(student.ID(), c.ID())
	require.NoError(t, enrollRepo.Create(ctx, e))

	chapterRepo := postgres.NewChapterRepository(pool)
	ch, _ := curriculum.NewChapter(c.ID(), "Chuong 1", 0)
	require.NoError(t, chapterRepo.Create(ctx, ch))

	lessonRepo := postgres.NewLessonRepository(pool)
	l, _ := curriculum.NewLesson(ch.ID(), "Bai 1", 0)
	require.NoError(t, lessonRepo.Create(ctx, l))

	assignmentRepo := postgres.NewAssignmentRepository(pool)
	quiz, _ := assignment.NewAssignment(l.ID(), "Trac nghiem", "", assignment.TypeQuiz, []assignment.Question{
		{Content: "1+1=?", Options: []string{"1", "2", "3"}, CorrectIndex: 1},
	}, nil)
	require.NoError(t, assignmentRepo.Create(ctx, quiz))

	submissionRepo := postgres.NewSubmissionRepository(pool)
	s, _ := submission.NewSubmission(quiz.ID(), student.ID(), "", []int{1})
	require.NoError(t, s.Grade(10, ""))
	require.NoError(t, submissionRepo.Create(ctx, s))

	gradebookRepo := postgres.NewGradebookRepository(pool)
	entries, err := gradebookRepo.ForCourse(ctx, c.ID())
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, student.ID(), entries[0].StudentID)
	assert.Equal(t, "Hoc Vien Nam", entries[0].StudentName)
	assert.Equal(t, quiz.ID(), entries[0].AssignmentID)
	require.NotNil(t, entries[0].Score)
	assert.Equal(t, 10.0, *entries[0].Score)
}
