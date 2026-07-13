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

func TestProgressRepository_ForStudent(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, _ := user.NewUser("gv8@vlu.edu.vn", "GV", user.RoleTeacher)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	student, _ := user.NewUser("hv8@vlu.edu.vn", "Hoc Vien", user.RoleStudent)
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
	a1, _ := assignment.NewAssignment(l.ID(), "Bai tap 1", "", assignment.TypeEssay, nil, nil, nil)
	require.NoError(t, assignmentRepo.Create(ctx, a1))
	a2, _ := assignment.NewAssignment(l.ID(), "Bai tap 2", "", assignment.TypeEssay, nil, nil, nil)
	require.NoError(t, assignmentRepo.Create(ctx, a2))

	submissionRepo := postgres.NewSubmissionRepository(pool)
	s, _ := submission.NewSubmission(a1.ID(), student.ID(), "bai lam", nil)
	require.NoError(t, submissionRepo.Create(ctx, s))

	progressRepo := postgres.NewProgressRepository(pool)
	list, err := progressRepo.ForStudent(ctx, student.ID())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, c.ID(), list[0].CourseID)
	assert.Equal(t, 2, list[0].TotalAssignments)
	assert.Equal(t, 1, list[0].Submitted)
	assert.Equal(t, 50.0, list[0].PercentComplete)
}
