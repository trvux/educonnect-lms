package curriculum_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/curriculum"
	curriculumservice "educonnect-lms/backend/internal/service/curriculum"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeChapterRepo struct {
	byID   map[uint]*curriculum.Chapter
	nextID uint
}

func newFakeChapterRepo() *fakeChapterRepo {
	return &fakeChapterRepo{byID: map[uint]*curriculum.Chapter{}}
}

func (r *fakeChapterRepo) Create(_ context.Context, c *curriculum.Chapter) error {
	r.nextID++
	c.SetID(r.nextID)
	r.byID[c.ID()] = c
	return nil
}
func (r *fakeChapterRepo) FindByID(_ context.Context, id uint) (*curriculum.Chapter, error) {
	if c, ok := r.byID[id]; ok {
		return c, nil
	}
	return nil, curriculum.ErrChapterNotFound
}
func (r *fakeChapterRepo) ListByCourse(_ context.Context, courseID uint) ([]*curriculum.Chapter, error) {
	var out []*curriculum.Chapter
	for _, c := range r.byID {
		if c.CourseID() == courseID {
			out = append(out, c)
		}
	}
	return out, nil
}
func (r *fakeChapterRepo) CountByCourse(_ context.Context, courseID uint) (int, error) {
	n := 0
	for _, c := range r.byID {
		if c.CourseID() == courseID {
			n++
		}
	}
	return n, nil
}
func (r *fakeChapterRepo) Update(_ context.Context, c *curriculum.Chapter) error {
	r.byID[c.ID()] = c
	return nil
}

type fakeLessonRepo struct {
	byID   map[uint]*curriculum.Lesson
	nextID uint
}

func newFakeLessonRepo() *fakeLessonRepo { return &fakeLessonRepo{byID: map[uint]*curriculum.Lesson{}} }

func (r *fakeLessonRepo) Create(_ context.Context, l *curriculum.Lesson) error {
	r.nextID++
	l.SetID(r.nextID)
	r.byID[l.ID()] = l
	return nil
}
func (r *fakeLessonRepo) FindByID(_ context.Context, id uint) (*curriculum.Lesson, error) {
	if l, ok := r.byID[id]; ok {
		return l, nil
	}
	return nil, curriculum.ErrLessonNotFound
}
func (r *fakeLessonRepo) ListByChapter(_ context.Context, chapterID uint) ([]*curriculum.Lesson, error) {
	var out []*curriculum.Lesson
	for _, l := range r.byID {
		if l.ChapterID() == chapterID {
			out = append(out, l)
		}
	}
	return out, nil
}
func (r *fakeLessonRepo) CountByChapter(_ context.Context, chapterID uint) (int, error) {
	n := 0
	for _, l := range r.byID {
		if l.ChapterID() == chapterID {
			n++
		}
	}
	return n, nil
}
func (r *fakeLessonRepo) Update(_ context.Context, l *curriculum.Lesson) error {
	r.byID[l.ID()] = l
	return nil
}

func TestService_CreateChapter_AutoPosition(t *testing.T) {
	s := curriculumservice.NewService(newFakeChapterRepo(), newFakeLessonRepo())
	ctx := context.Background()

	c1, err := s.CreateChapter(ctx, 1, "Chuong 1")
	require.NoError(t, err)
	assert.Equal(t, 0, c1.Position())

	c2, err := s.CreateChapter(ctx, 1, "Chuong 2")
	require.NoError(t, err)
	assert.Equal(t, 1, c2.Position(), "chapter thu 2 phai tu dong xep sau chapter 1")
}

func TestService_CreateLesson_RequiresExistingChapter(t *testing.T) {
	s := curriculumservice.NewService(newFakeChapterRepo(), newFakeLessonRepo())
	ctx := context.Background()

	_, err := s.CreateLesson(ctx, 999, "Bai 1")
	assert.ErrorIs(t, err, curriculum.ErrChapterNotFound)

	ch, err := s.CreateChapter(ctx, 1, "Chuong 1")
	require.NoError(t, err)

	l, err := s.CreateLesson(ctx, ch.ID(), "Bai 1")
	require.NoError(t, err)
	assert.Equal(t, 0, l.Position())
}
