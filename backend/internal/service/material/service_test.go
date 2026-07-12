package material_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/material"
	materialservice "educonnect-lms/backend/internal/service/material"

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

type fakeMaterialRepo struct {
	items  []*material.Material
	nextID uint
}

func (r *fakeMaterialRepo) Create(_ context.Context, m *material.Material) error {
	r.nextID++
	m.SetID(r.nextID)
	r.items = append(r.items, m)
	return nil
}
func (r *fakeMaterialRepo) ListByLesson(_ context.Context, lessonID uint) ([]*material.Material, error) {
	var out []*material.Material
	for _, m := range r.items {
		if m.LessonID() == lessonID {
			out = append(out, m)
		}
	}
	return out, nil
}

type fakeStorage struct{ savedPath string }

func (s *fakeStorage) Save(_ context.Context, subdir, fileName string, content io.Reader) (string, error) {
	_, _ = io.ReadAll(content) // giả lập ghi file, không cần lưu thật trong test
	s.savedPath = subdir + "/" + fileName
	return s.savedPath, nil
}

func TestService_Upload(t *testing.T) {
	lessons := &fakeLessonRepo{exists: map[uint]bool{1: true}}
	materials := &fakeMaterialRepo{}
	storage := &fakeStorage{}
	svc := materialservice.NewService(materials, lessons, storage)

	m, err := svc.Upload(context.Background(), 1, "slide-bai-1.pdf", bytes.NewBufferString("noi dung gia"))
	require.NoError(t, err)
	assert.Equal(t, "slide-bai-1.pdf", m.FileName())
	assert.Equal(t, "lesson-1/slide-bai-1.pdf", m.FilePath())

	list, err := svc.ListByLesson(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestService_Upload_LessonNotFound(t *testing.T) {
	lessons := &fakeLessonRepo{exists: map[uint]bool{}}
	svc := materialservice.NewService(&fakeMaterialRepo{}, lessons, &fakeStorage{})

	_, err := svc.Upload(context.Background(), 999, "a.pdf", bytes.NewBufferString("x"))
	assert.ErrorIs(t, err, curriculum.ErrLessonNotFound)
}
