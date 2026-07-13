package submission_test

import (
	"context"
	"testing"
	"time"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/quizattempt"
	"educonnect-lms/backend/internal/domain/submission"
	submissionservice "educonnect-lms/backend/internal/service/submission"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAssignmentGetter struct {
	items map[uint]*assignment.Assignment
}

func (r *fakeAssignmentGetter) Get(_ context.Context, id uint) (*assignment.Assignment, error) {
	a, ok := r.items[id]
	if !ok {
		return nil, assignment.ErrNotFound
	}
	return a, nil
}

type fakeSubmissionRepo struct {
	items  []*submission.Submission
	nextID uint
}

func (r *fakeSubmissionRepo) Create(_ context.Context, s *submission.Submission) error {
	r.nextID++
	s.SetID(r.nextID)
	r.items = append(r.items, s)
	return nil
}
func (r *fakeSubmissionRepo) FindByID(_ context.Context, id uint) (*submission.Submission, error) {
	for _, s := range r.items {
		if s.ID() == id {
			return s, nil
		}
	}
	return nil, submission.ErrNotFound
}
func (r *fakeSubmissionRepo) Update(_ context.Context, updated *submission.Submission) error {
	for _, s := range r.items {
		if s.ID() == updated.ID() {
			return nil
		}
	}
	return submission.ErrNotFound
}
func (r *fakeSubmissionRepo) FindByAssignmentAndStudent(_ context.Context, assignmentID, studentID uint) (*submission.Submission, error) {
	for _, s := range r.items {
		if s.AssignmentID() == assignmentID && s.StudentID() == studentID {
			return s, nil
		}
	}
	return nil, submission.ErrNotFound
}
func (r *fakeSubmissionRepo) ListByAssignment(_ context.Context, assignmentID uint) ([]*submission.Submission, error) {
	var out []*submission.Submission
	for _, s := range r.items {
		if s.AssignmentID() == assignmentID {
			out = append(out, s)
		}
	}
	return out, nil
}

type fakeLessonRepo struct{ items map[uint]*curriculum.Lesson }

func (r *fakeLessonRepo) Create(_ context.Context, _ *curriculum.Lesson) error { return nil }
func (r *fakeLessonRepo) FindByID(_ context.Context, id uint) (*curriculum.Lesson, error) {
	l, ok := r.items[id]
	if !ok {
		return nil, curriculum.ErrLessonNotFound
	}
	return l, nil
}
func (r *fakeLessonRepo) ListByChapter(_ context.Context, _ uint) ([]*curriculum.Lesson, error) {
	return nil, nil
}
func (r *fakeLessonRepo) CountByChapter(_ context.Context, _ uint) (int, error) { return 0, nil }
func (r *fakeLessonRepo) Update(_ context.Context, _ *curriculum.Lesson) error  { return nil }
func (r *fakeLessonRepo) Delete(_ context.Context, _ uint) error                { return nil }

type fakeChapterRepo struct{ items map[uint]*curriculum.Chapter }

func (r *fakeChapterRepo) Create(_ context.Context, _ *curriculum.Chapter) error { return nil }
func (r *fakeChapterRepo) FindByID(_ context.Context, id uint) (*curriculum.Chapter, error) {
	c, ok := r.items[id]
	if !ok {
		return nil, curriculum.ErrChapterNotFound
	}
	return c, nil
}
func (r *fakeChapterRepo) ListByCourse(_ context.Context, _ uint) ([]*curriculum.Chapter, error) {
	return nil, nil
}
func (r *fakeChapterRepo) CountByCourse(_ context.Context, _ uint) (int, error)  { return 0, nil }
func (r *fakeChapterRepo) Update(_ context.Context, _ *curriculum.Chapter) error { return nil }
func (r *fakeChapterRepo) Delete(_ context.Context, _ uint) error                { return nil }

type fakeCourseGetter struct{ items map[uint]*course.Course }

func (r *fakeCourseGetter) FindByID(_ context.Context, id uint) (*course.Course, error) {
	c, ok := r.items[id]
	if !ok {
		return nil, course.ErrNotFound
	}
	return c, nil
}

// fakeQuizAttemptGetter (US5.4) — chỉ hiện thực method service cần
// (submission.QuizAttemptGetter), preset sẵn attempt cho từng test case.
type fakeQuizAttemptGetter struct {
	items map[uint]*quizattempt.QuizAttempt // assignmentID -> attempt (test chỉ dùng 1 student)
}

func (r *fakeQuizAttemptGetter) FindByAssignmentAndStudent(_ context.Context, assignmentID, _ uint) (*quizattempt.QuizAttempt, error) {
	a, ok := r.items[assignmentID]
	if !ok {
		return nil, quizattempt.ErrNotFound
	}
	return a, nil
}

func noAttempts() *fakeQuizAttemptGetter {
	return &fakeQuizAttemptGetter{items: map[uint]*quizattempt.QuizAttempt{}}
}

// newTestSetup dựng chuỗi sở hữu Course(teacherID) → Chapter → Lesson →
// Assignment(kind, dueAt, questions, timeLimitMinutes) hoàn chỉnh, dùng
// chung cho các test Submit/Grade cần xác minh quyền sở hữu (US5.3) và giới
// hạn thời gian (US5.4).
func newTestSetup(teacherID uint, kind assignment.Type, dueAt *time.Time, timeLimitMinutes *int) (*fakeAssignmentGetter, *fakeLessonRepo, *fakeChapterRepo, *fakeCourseGetter) {
	c, _ := course.NewCourse("Khoa hoc", "d", teacherID)
	c.SetID(1)

	ch, _ := curriculum.NewChapter(c.ID(), "Chuong 1", 0)
	ch.SetID(1)

	l, _ := curriculum.NewLesson(ch.ID(), "Bai 1", 0)
	l.SetID(1)

	var questions []assignment.Question
	if kind == assignment.TypeQuiz {
		questions = []assignment.Question{
			{Content: "1+1=?", Options: []string{"1", "2", "3"}, CorrectIndex: 1},
			{Content: "2+2=?", Options: []string{"3", "4", "5"}, CorrectIndex: 1},
		}
	}
	a, _ := assignment.NewAssignment(l.ID(), "Bai tap", "", kind, questions, dueAt, timeLimitMinutes)
	a.SetID(1)

	return &fakeAssignmentGetter{items: map[uint]*assignment.Assignment{1: a}},
		&fakeLessonRepo{items: map[uint]*curriculum.Lesson{1: l}},
		&fakeChapterRepo{items: map[uint]*curriculum.Chapter{1: ch}},
		&fakeCourseGetter{items: map[uint]*course.Course{1: c}}
}

func TestService_Submit_QuizAutoGraded(t *testing.T) {
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeQuiz, nil, nil)
	svc := submissionservice.NewService(&fakeSubmissionRepo{}, assignments, lessons, chapters, courses, noAttempts())

	// 1/2 câu đúng -> 5.0 điểm
	s, err := svc.Submit(context.Background(), 1, 2, "", []int{1, 0})
	require.NoError(t, err)
	require.True(t, s.IsGraded())
	assert.Equal(t, 5.0, *s.Score())
}

func TestService_Submit_AssignmentNotFound(t *testing.T) {
	svc := submissionservice.NewService(&fakeSubmissionRepo{}, &fakeAssignmentGetter{items: map[uint]*assignment.Assignment{}}, &fakeLessonRepo{}, &fakeChapterRepo{}, &fakeCourseGetter{}, noAttempts())

	_, err := svc.Submit(context.Background(), 999, 2, "", []int{1})
	assert.ErrorIs(t, err, assignment.ErrNotFound)
}

func TestService_Submit_PastDue(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour)
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeQuiz, &past, nil)
	svc := submissionservice.NewService(&fakeSubmissionRepo{}, assignments, lessons, chapters, courses, noAttempts())

	_, err := svc.Submit(context.Background(), 1, 2, "", []int{1, 1})
	assert.ErrorIs(t, err, submission.ErrPastDue)
}

func TestService_Submit_AnswerCountMismatch(t *testing.T) {
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeQuiz, nil, nil)
	svc := submissionservice.NewService(&fakeSubmissionRepo{}, assignments, lessons, chapters, courses, noAttempts())

	_, err := svc.Submit(context.Background(), 1, 2, "", []int{1})
	assert.ErrorIs(t, err, submission.ErrAnswerCountMismatch)
}

func TestService_Submit_AlreadySubmitted(t *testing.T) {
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeQuiz, nil, nil)
	submissions := &fakeSubmissionRepo{}
	svc := submissionservice.NewService(submissions, assignments, lessons, chapters, courses, noAttempts())

	_, err := svc.Submit(context.Background(), 1, 2, "", []int{1, 1})
	require.NoError(t, err)

	_, err = svc.Submit(context.Background(), 1, 2, "", []int{0, 0})
	assert.ErrorIs(t, err, submission.ErrAlreadySubmitted)
}

// US5.4
func TestService_Submit_TimeExpired(t *testing.T) {
	limit := 10 // phút
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeQuiz, nil, &limit)
	startedLongAgo := quizattempt.Rehydrate(1, 1, 2, time.Now().UTC().Add(-30*time.Minute))
	attempts := &fakeQuizAttemptGetter{items: map[uint]*quizattempt.QuizAttempt{1: startedLongAgo}}
	svc := submissionservice.NewService(&fakeSubmissionRepo{}, assignments, lessons, chapters, courses, attempts)

	_, err := svc.Submit(context.Background(), 1, 2, "", []int{1, 1})
	assert.ErrorIs(t, err, submission.ErrTimeExpired)
}

// US5.4 — vẫn trong giới hạn thời gian thì nộp bình thường.
func TestService_Submit_WithinTimeLimit_Allowed(t *testing.T) {
	limit := 10
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeQuiz, nil, &limit)
	startedRecently := quizattempt.Rehydrate(1, 1, 2, time.Now().UTC().Add(-2*time.Minute))
	attempts := &fakeQuizAttemptGetter{items: map[uint]*quizattempt.QuizAttempt{1: startedRecently}}
	svc := submissionservice.NewService(&fakeSubmissionRepo{}, assignments, lessons, chapters, courses, attempts)

	_, err := svc.Submit(context.Background(), 1, 2, "", []int{1, 1})
	assert.NoError(t, err)
}

// US5.4 — chưa từng "bắt đầu" attempt (gọi thẳng API submit) thì không đủ
// dữ kiện để tính hết giờ, không phạt oan.
func TestService_Submit_TimeLimit_NoAttempt_Allowed(t *testing.T) {
	limit := 10
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeQuiz, nil, &limit)
	svc := submissionservice.NewService(&fakeSubmissionRepo{}, assignments, lessons, chapters, courses, noAttempts())

	_, err := svc.Submit(context.Background(), 1, 2, "", []int{1, 1})
	assert.NoError(t, err)
}

func TestService_Grade_Essay_ByOwningTeacher(t *testing.T) {
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeEssay, nil, nil)
	submissions := &fakeSubmissionRepo{}
	svc := submissionservice.NewService(submissions, assignments, lessons, chapters, courses, noAttempts())

	sub, err := svc.Submit(context.Background(), 1, 2, "bai lam tu luan", nil)
	require.NoError(t, err)
	assert.False(t, sub.IsGraded())

	graded, err := svc.Grade(context.Background(), sub.ID(), 10, false, 8.5, "Lam tot")
	require.NoError(t, err)
	assert.True(t, graded.IsGraded())
	assert.Equal(t, 8.5, *graded.Score())
}

func TestService_Grade_ByOtherTeacher_Forbidden(t *testing.T) {
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeEssay, nil, nil)
	submissions := &fakeSubmissionRepo{}
	svc := submissionservice.NewService(submissions, assignments, lessons, chapters, courses, noAttempts())

	sub, err := svc.Submit(context.Background(), 1, 2, "bai lam", nil)
	require.NoError(t, err)

	_, err = svc.Grade(context.Background(), sub.ID(), 999, false, 8, "")
	assert.ErrorIs(t, err, submission.ErrNotFound)
}

func TestService_Grade_ByAdmin_Allowed(t *testing.T) {
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeEssay, nil, nil)
	submissions := &fakeSubmissionRepo{}
	svc := submissionservice.NewService(submissions, assignments, lessons, chapters, courses, noAttempts())

	sub, err := svc.Submit(context.Background(), 1, 2, "bai lam", nil)
	require.NoError(t, err)

	graded, err := svc.Grade(context.Background(), sub.ID(), 999, true, 9, "")
	require.NoError(t, err)
	assert.Equal(t, 9.0, *graded.Score())
}

func TestService_GetMine(t *testing.T) {
	assignments, lessons, chapters, courses := newTestSetup(10, assignment.TypeEssay, nil, nil)
	submissions := &fakeSubmissionRepo{}
	svc := submissionservice.NewService(submissions, assignments, lessons, chapters, courses, noAttempts())

	_, err := svc.GetMine(context.Background(), 1, 2)
	assert.ErrorIs(t, err, submission.ErrNotFound)

	sub, err := svc.Submit(context.Background(), 1, 2, "bai lam", nil)
	require.NoError(t, err)

	mine, err := svc.GetMine(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, sub.ID(), mine.ID())
}
