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

func TestReportRepository(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	teacher, _ := user.NewUser("gv9@vlu.edu.vn", "GV", user.RoleTeacher)
	require.NoError(t, teacher.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, teacher))

	student, _ := user.NewUser("hv9@vlu.edu.vn", "Hoc Vien", user.RoleStudent)
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

	reportRepo := postgres.NewReportRepository(pool)

	forTeacher, err := reportRepo.ForTeacher(ctx, teacher.ID())
	require.NoError(t, err)
	require.Len(t, forTeacher, 1)
	assert.Equal(t, 1, forTeacher[0].EnrolledStudents)
	assert.Equal(t, 2, forTeacher[0].TotalAssignments)
	assert.Equal(t, 50.0, forTeacher[0].AverageCompletion)

	all, err := reportRepo.All(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, all)
}
