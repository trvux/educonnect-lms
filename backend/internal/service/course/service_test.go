package course_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	courseservice "educonnect-lms/backend/internal/service/course"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCourseRepo struct {
	byID   map[uint]*course.Course
	nextID uint
}

func newFakeCourseRepo() *fakeCourseRepo {
	return &fakeCourseRepo{byID: map[uint]*course.Course{}}
}

func (r *fakeCourseRepo) Create(_ context.Context, c *course.Course) error {
	r.nextID++
	c.SetID(r.nextID)
	r.byID[c.ID()] = c
	return nil
}

func (r *fakeCourseRepo) FindByID(_ context.Context, id uint) (*course.Course, error) {
	if c, ok := r.byID[id]; ok {
		return c, nil
	}
	return nil, course.ErrNotFound
}

func (r *fakeCourseRepo) Search(_ context.Context, keyword string) ([]*course.Course, error) {
	var out []*course.Course
	for _, c := range r.byID {
		if c.IsSearchable() {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *fakeCourseRepo) ListByTeacher(_ context.Context, teacherID uint) ([]*course.Course, error) {
	var out []*course.Course
	for _, c := range r.byID {
		if c.TeacherID() == teacherID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *fakeCourseRepo) Update(_ context.Context, c *course.Course) error {
	r.byID[c.ID()] = c
	return nil
}

func TestService_Create(t *testing.T) {
	repo := newFakeCourseRepo()
	s := courseservice.NewService(repo)

	c, err := s.Create(context.Background(), courseservice.CreateInput{
		Title: "Nhap mon Golang", Description: "desc", TeacherID: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, course.StatusDraft, c.Status())
}

func TestService_SubmitAndApprove(t *testing.T) {
	repo := newFakeCourseRepo()
	s := courseservice.NewService(repo)
	ctx := context.Background()

	c, err := s.Create(ctx, courseservice.CreateInput{Title: "Golang", TeacherID: 1})
	require.NoError(t, err)

	// Giáo viên khác không được submit khóa học không phải của mình.
	_, err = s.SubmitForReview(ctx, c.ID(), 999)
	assert.ErrorIs(t, err, course.ErrNotFound)

	_, err = s.SubmitForReview(ctx, c.ID(), 1)
	require.NoError(t, err)

	approved, err := s.Approve(ctx, c.ID())
	require.NoError(t, err)
	assert.True(t, approved.IsSearchable())
}

func TestService_Search_OnlyApproved(t *testing.T) {
	repo := newFakeCourseRepo()
	s := courseservice.NewService(repo)
	ctx := context.Background()

	draft, err := s.Create(ctx, courseservice.CreateInput{Title: "Draft course", TeacherID: 1})
	require.NoError(t, err)

	results, err := s.Search(ctx, "")
	require.NoError(t, err)
	assert.Empty(t, results, "draft course must not appear in search")

	_, err = s.SubmitForReview(ctx, draft.ID(), 1)
	require.NoError(t, err)
	_, err = s.Approve(ctx, draft.ID())
	require.NoError(t, err)

	results, err = s.Search(ctx, "")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}
