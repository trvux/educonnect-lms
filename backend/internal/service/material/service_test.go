package material_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/material"
	"educonnect-lms/backend/internal/domain/user"
	materialservice "educonnect-lms/backend/internal/service/material"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

type fakeEnrollmentChecker struct{ enrolled map[uint]map[uint]bool } // studentID -> courseID -> bool

func (r *fakeEnrollmentChecker) IsEnrolled(_ context.Context, studentID, courseID uint) (bool, error) {
	return r.enrolled[studentID][courseID], nil
}

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
func (r *fakeMaterialRepo) FindByID(_ context.Context, id uint) (*material.Material, error) {
	for _, m := range r.items {
		if m.ID() == id {
			return m, nil
		}
	}
	return nil, material.ErrNotFound
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
func (r *fakeMaterialRepo) Delete(_ context.Context, id uint) error {
	for i, m := range r.items {
		if m.ID() == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return material.ErrNotFound
}

type fakeStorage struct {
	savedPath   string
	deletedPath string
}

func (s *fakeStorage) Save(_ context.Context, subdir, fileName string, content io.Reader) (string, error) {
	_, _ = io.ReadAll(content) // giả lập ghi file, không cần lưu thật trong test
	s.savedPath = subdir + "/" + fileName
	return s.savedPath, nil
}
func (s *fakeStorage) Delete(_ context.Context, path string) error {
	s.deletedPath = path
	return nil
}

// setupCourse dựng course(teacherID=1) -> chapter -> lesson, kèm 1 học viên
// đã đăng ký (id=2) và chưa đăng ký (id=3), dùng chung cho các test US4.3.
func setupCourse(t *testing.T) (lessons *fakeLessonRepo, chapters *fakeChapterRepo, courses *fakeCourseGetter, enrollments *fakeEnrollmentChecker, lessonID uint) {
	t.Helper()
	c, err := course.NewCourse("Golang", "desc", 1)
	require.NoError(t, err)
	c.SetID(10)

	ch, err := curriculum.NewChapter(c.ID(), "Chuong 1", 0)
	require.NoError(t, err)
	ch.SetID(20)

	l, err := curriculum.NewLesson(ch.ID(), "Bai 1", 0)
	require.NoError(t, err)
	l.SetID(30)

	return &fakeLessonRepo{items: map[uint]*curriculum.Lesson{l.ID(): l}},
		&fakeChapterRepo{items: map[uint]*curriculum.Chapter{ch.ID(): ch}},
		&fakeCourseGetter{items: map[uint]*course.Course{c.ID(): c}},
		&fakeEnrollmentChecker{enrolled: map[uint]map[uint]bool{2: {c.ID(): true}}},
		l.ID()
}

func TestService_Upload(t *testing.T) {
	lessons := &fakeLessonRepo{items: map[uint]*curriculum.Lesson{1: mustLesson(t, 1)}}
	materials := &fakeMaterialRepo{}
	storage := &fakeStorage{}
	svc := materialservice.NewService(materials, lessons, &fakeChapterRepo{}, &fakeCourseGetter{}, &fakeEnrollmentChecker{}, storage)

	m, err := svc.Upload(context.Background(), 1, "slide-bai-1.pdf", bytes.NewBufferString("noi dung gia"))
	require.NoError(t, err)
	assert.Equal(t, "slide-bai-1.pdf", m.FileName())
	assert.Equal(t, "lesson-1/slide-bai-1.pdf", m.FilePath())
}

func TestService_Upload_LessonNotFound(t *testing.T) {
	lessons := &fakeLessonRepo{items: map[uint]*curriculum.Lesson{}}
	svc := materialservice.NewService(&fakeMaterialRepo{}, lessons, &fakeChapterRepo{}, &fakeCourseGetter{}, &fakeEnrollmentChecker{}, &fakeStorage{})

	_, err := svc.Upload(context.Background(), 999, "a.pdf", bytes.NewBufferString("x"))
	assert.ErrorIs(t, err, curriculum.ErrLessonNotFound)
}

func TestService_ListByLesson_AccessControl(t *testing.T) {
	lessons, chapters, courses, enrollments, lessonID := setupCourse(t)
	materials := &fakeMaterialRepo{}
	svc := materialservice.NewService(materials, lessons, chapters, courses, enrollments, &fakeStorage{})
	_, err := svc.Upload(context.Background(), lessonID, "slide.pdf", bytes.NewBufferString("x"))
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("admin luôn được phép", func(t *testing.T) {
		list, err := svc.ListByLesson(ctx, lessonID, 999, user.RoleAdmin)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("giảng viên sở hữu khóa học được phép", func(t *testing.T) {
		list, err := svc.ListByLesson(ctx, lessonID, 1, user.RoleTeacher)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("giảng viên không sở hữu khóa học bị chặn", func(t *testing.T) {
		_, err := svc.ListByLesson(ctx, lessonID, 999, user.RoleTeacher)
		assert.ErrorIs(t, err, material.ErrNotFound)
	})

	t.Run("học viên đã đăng ký được phép", func(t *testing.T) {
		list, err := svc.ListByLesson(ctx, lessonID, 2, user.RoleStudent)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("học viên chưa đăng ký bị chặn", func(t *testing.T) {
		_, err := svc.ListByLesson(ctx, lessonID, 3, user.RoleStudent)
		assert.ErrorIs(t, err, material.ErrNotFound)
	})
}

func TestService_Get_AccessControl(t *testing.T) {
	lessons, chapters, courses, enrollments, lessonID := setupCourse(t)
	materials := &fakeMaterialRepo{}
	svc := materialservice.NewService(materials, lessons, chapters, courses, enrollments, &fakeStorage{})
	m, err := svc.Upload(context.Background(), lessonID, "slide.pdf", bytes.NewBufferString("x"))
	require.NoError(t, err)
	ctx := context.Background()

	_, err = svc.Get(ctx, m.ID(), 2, user.RoleStudent)
	require.NoError(t, err, "học viên đã đăng ký phải tải được")

	_, err = svc.Get(ctx, m.ID(), 3, user.RoleStudent)
	assert.ErrorIs(t, err, material.ErrNotFound, "học viên chưa đăng ký không được tải")
}

func TestService_Delete_ManagerOnly(t *testing.T) {
	lessons, chapters, courses, enrollments, lessonID := setupCourse(t)
	materials := &fakeMaterialRepo{}
	storage := &fakeStorage{}
	svc := materialservice.NewService(materials, lessons, chapters, courses, enrollments, storage)
	ctx := context.Background()

	m, err := svc.Upload(ctx, lessonID, "slide.pdf", bytes.NewBufferString("x"))
	require.NoError(t, err)

	// Học viên đã đăng ký (userID=2) KHÔNG được xóa dù xem/tải được (US4.8).
	err = svc.Delete(ctx, m.ID(), 2, user.RoleStudent)
	assert.ErrorIs(t, err, material.ErrNotFound, "học viên không bao giờ được xóa tài liệu")

	// Giảng viên không sở hữu khóa học không được xóa.
	err = svc.Delete(ctx, m.ID(), 999, user.RoleTeacher)
	assert.ErrorIs(t, err, material.ErrNotFound)

	// Giảng viên sở hữu khóa học xóa được — cả file vật lý lẫn metadata.
	require.NoError(t, svc.Delete(ctx, m.ID(), 1, user.RoleTeacher))
	assert.Equal(t, m.FilePath(), storage.deletedPath, "phải gọi storage.Delete đúng đường dẫn file")

	_, err = materials.FindByID(ctx, m.ID())
	assert.ErrorIs(t, err, material.ErrNotFound, "metadata phải bị xóa khỏi repository")
}

func TestService_Delete_AdminAlwaysAllowed(t *testing.T) {
	lessons, chapters, courses, enrollments, lessonID := setupCourse(t)
	materials := &fakeMaterialRepo{}
	svc := materialservice.NewService(materials, lessons, chapters, courses, enrollments, &fakeStorage{})
	ctx := context.Background()

	m, err := svc.Upload(ctx, lessonID, "slide.pdf", bytes.NewBufferString("x"))
	require.NoError(t, err)

	require.NoError(t, svc.Delete(ctx, m.ID(), 999, user.RoleAdmin))
}

func mustLesson(t *testing.T, id uint) *curriculum.Lesson {
	t.Helper()
	l, err := curriculum.NewLesson(1, "Bai 1", 0)
	require.NoError(t, err)
	l.SetID(id)
	return l
}
