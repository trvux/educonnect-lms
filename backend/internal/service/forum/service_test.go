package forum_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/forum"
	forumservice "educonnect-lms/backend/internal/service/forum"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCourseGetter struct{ items map[uint]*course.Course }

func (r *fakeCourseGetter) FindByID(_ context.Context, id uint) (*course.Course, error) {
	c, ok := r.items[id]
	if !ok {
		return nil, course.ErrNotFound
	}
	return c, nil
}

type fakePostRepo struct {
	items  []*forum.Post
	nextID uint
}

func (r *fakePostRepo) Create(_ context.Context, p *forum.Post) error {
	r.nextID++
	p.SetID(r.nextID)
	r.items = append(r.items, p)
	return nil
}
func (r *fakePostRepo) FindByID(_ context.Context, id uint) (*forum.Post, error) {
	for _, p := range r.items {
		if p.ID() == id {
			return p, nil
		}
	}
	return nil, forum.ErrNotFound
}
func (r *fakePostRepo) ListByCourse(_ context.Context, courseID uint) ([]*forum.Post, error) {
	var out []*forum.Post
	for _, p := range r.items {
		if p.CourseID() == courseID {
			out = append(out, p)
		}
	}
	return out, nil
}

func newCourse(id uint) *course.Course {
	c, _ := course.NewCourse("Golang", "d", 1)
	c.SetID(id)
	return c
}

func TestService_Post_Question(t *testing.T) {
	courses := &fakeCourseGetter{items: map[uint]*course.Course{1: newCourse(1)}}
	posts := &fakePostRepo{}
	svc := forumservice.NewService(posts, courses)

	p, err := svc.Post(context.Background(), 1, 2, nil, "Cau hoi cua toi")
	require.NoError(t, err)
	assert.NotZero(t, p.ID())
	assert.Nil(t, p.ParentID())
}

func TestService_Post_Reply(t *testing.T) {
	courses := &fakeCourseGetter{items: map[uint]*course.Course{1: newCourse(1)}}
	posts := &fakePostRepo{}
	svc := forumservice.NewService(posts, courses)

	question, err := svc.Post(context.Background(), 1, 2, nil, "Cau hoi")
	require.NoError(t, err)

	parentID := question.ID()
	reply, err := svc.Post(context.Background(), 1, 3, &parentID, "Cau tra loi")
	require.NoError(t, err)
	require.NotNil(t, reply.ParentID())
	assert.Equal(t, question.ID(), *reply.ParentID())
}

func TestService_Post_CourseNotFound(t *testing.T) {
	svc := forumservice.NewService(&fakePostRepo{}, &fakeCourseGetter{items: map[uint]*course.Course{}})

	_, err := svc.Post(context.Background(), 999, 2, nil, "noi dung")
	assert.ErrorIs(t, err, course.ErrNotFound)
}

func TestService_Post_ParentCourseMismatch(t *testing.T) {
	courses := &fakeCourseGetter{items: map[uint]*course.Course{1: newCourse(1), 2: newCourse(2)}}
	posts := &fakePostRepo{}
	svc := forumservice.NewService(posts, courses)

	question, err := svc.Post(context.Background(), 1, 2, nil, "Cau hoi khoa 1")
	require.NoError(t, err)

	parentID := question.ID()
	_, err = svc.Post(context.Background(), 2, 3, &parentID, "Tra loi sai khoa hoc")
	assert.ErrorIs(t, err, forum.ErrParentCourseMismatch)
}

func TestService_ListByCourse(t *testing.T) {
	courses := &fakeCourseGetter{items: map[uint]*course.Course{1: newCourse(1)}}
	posts := &fakePostRepo{}
	svc := forumservice.NewService(posts, courses)

	_, err := svc.Post(context.Background(), 1, 2, nil, "Cau hoi 1")
	require.NoError(t, err)

	list, err := svc.ListByCourse(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}
