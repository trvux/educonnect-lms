package assignment_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/curriculum"
	assignmentservice "educonnect-lms/backend/internal/service/assignment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeLessonRepo struct{ exists map[uint]bool }

func (r *fakeLessonRepo) Create(_ context.Context, _ *curriculum.Lesson) error { return nil }
func (r *fakeLessonRepo) FindByID(_ context.Context, id uint) (*curriculum.Lesson, error) {
	if r.exists[id] {
		l, _ := curriculum.NewLesson(1, "Bai 1", 0)
		l.SetID(id)
		return l, nil
	}
	return nil, curriculum.ErrLessonNotFound
}
func (r *fakeLessonRepo) ListByChapter(_ context.Context, _ uint) ([]*curriculum.Lesson, error) {
	return nil, nil
}
func (r *fakeLessonRepo) CountByChapter(_ context.Context, _ uint) (int, error) { return 0, nil }
func (r *fakeLessonRepo) Update(_ context.Context, _ *curriculum.Lesson) error  { return nil }
func (r *fakeLessonRepo) Delete(_ context.Context, _ uint) error                { return nil }

type fakeAssignmentRepo struct {
	items  []*assignment.Assignment
	nextID uint
}

func (r *fakeAssignmentRepo) Create(_ context.Context, a *assignment.Assignment) error {
	r.nextID++
	a.SetID(r.nextID)
	r.items = append(r.items, a)
	return nil
}
func (r *fakeAssignmentRepo) FindByID(_ context.Context, id uint) (*assignment.Assignment, error) {
	for _, a := range r.items {
		if a.ID() == id {
			return a, nil
		}
	}
	return nil, assignment.ErrNotFound
}
func (r *fakeAssignmentRepo) ListByLesson(_ context.Context, lessonID uint) ([]*assignment.Assignment, error) {
	var out []*assignment.Assignment
	for _, a := range r.items {
		if a.LessonID() == lessonID {
			out = append(out, a)
		}
	}
	return out, nil
}

func TestService_Create(t *testing.T) {
	lessons := &fakeLessonRepo{exists: map[uint]bool{1: true}}
	assignments := &fakeAssignmentRepo{}
	svc := assignmentservice.NewService(assignments, lessons)

	a, err := svc.Create(context.Background(), 1, "Bai tap tu luan", "Nop file", assignment.TypeEssay, nil, nil)
	require.NoError(t, err)
	assert.NotZero(t, a.ID())

	got, err := svc.Get(context.Background(), a.ID())
	require.NoError(t, err)
	assert.Equal(t, "Bai tap tu luan", got.Title())

	list, err := svc.ListByLesson(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestService_Create_LessonNotFound(t *testing.T) {
	lessons := &fakeLessonRepo{exists: map[uint]bool{}}
	svc := assignmentservice.NewService(&fakeAssignmentRepo{}, lessons)

	_, err := svc.Create(context.Background(), 999, "Bai tap", "", assignment.TypeEssay, nil, nil)
	assert.ErrorIs(t, err, curriculum.ErrLessonNotFound)
}

func TestService_Create_InvalidAssignment(t *testing.T) {
	lessons := &fakeLessonRepo{exists: map[uint]bool{1: true}}
	svc := assignmentservice.NewService(&fakeAssignmentRepo{}, lessons)

	_, err := svc.Create(context.Background(), 1, "Trac nghiem", "", assignment.TypeQuiz, nil, nil)
	assert.ErrorIs(t, err, assignment.ErrQuizNeedsQuestions)
}
