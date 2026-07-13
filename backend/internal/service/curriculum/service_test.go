package curriculum_test

import (
	"context"
	"sort"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/user"
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
	// Postgres thật có ORDER BY position ASC (US4.7 cần đúng thứ tự này).
	sort.Slice(out, func(i, j int) bool { return out[i].Position() < out[j].Position() })
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
func (r *fakeChapterRepo) Delete(_ context.Context, id uint) error {
	if _, ok := r.byID[id]; !ok {
		return curriculum.ErrChapterNotFound
	}
	delete(r.byID, id)
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
	sort.Slice(out, func(i, j int) bool { return out[i].Position() < out[j].Position() })
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
func (r *fakeLessonRepo) Delete(_ context.Context, id uint) error {
	if _, ok := r.byID[id]; !ok {
		return curriculum.ErrLessonNotFound
	}
	delete(r.byID, id)
	return nil
}

type fakeCourseGetter struct{ byID map[uint]*course.Course }

func newFakeCourseGetter() *fakeCourseGetter {
	return &fakeCourseGetter{byID: map[uint]*course.Course{}}
}

func (r *fakeCourseGetter) FindByID(_ context.Context, id uint) (*course.Course, error) {
	if c, ok := r.byID[id]; ok {
		return c, nil
	}
	return nil, course.ErrNotFound
}

func TestService_CreateChapter_AutoPosition(t *testing.T) {
	s := curriculumservice.NewService(newFakeChapterRepo(), newFakeLessonRepo(), newFakeCourseGetter())
	ctx := context.Background()

	c1, err := s.CreateChapter(ctx, 1, "Chuong 1")
	require.NoError(t, err)
	assert.Equal(t, 0, c1.Position())

	c2, err := s.CreateChapter(ctx, 1, "Chuong 2")
	require.NoError(t, err)
	assert.Equal(t, 1, c2.Position(), "chapter thu 2 phai tu dong xep sau chapter 1")
}

func TestService_CreateLesson_RequiresExistingChapter(t *testing.T) {
	s := curriculumservice.NewService(newFakeChapterRepo(), newFakeLessonRepo(), newFakeCourseGetter())
	ctx := context.Background()

	_, err := s.CreateLesson(ctx, 999, "Bai 1")
	assert.ErrorIs(t, err, curriculum.ErrChapterNotFound)

	ch, err := s.CreateChapter(ctx, 1, "Chuong 1")
	require.NoError(t, err)

	l, err := s.CreateLesson(ctx, ch.ID(), "Bai 1")
	require.NoError(t, err)
	assert.Equal(t, 0, l.Position())
}

// setupOwnership dựng course(teacherID=1) -> chapter -> lesson, dùng chung
// cho các test US4.6 (quyền sửa/xóa).
func setupOwnership(t *testing.T) (*curriculumservice.Service, uint, uint) {
	t.Helper()
	c, err := course.NewCourse("Golang", "desc", 1)
	require.NoError(t, err)
	c.SetID(10)

	courses := newFakeCourseGetter()
	courses.byID[c.ID()] = c

	chapters := newFakeChapterRepo()
	lessons := newFakeLessonRepo()
	s := curriculumservice.NewService(chapters, lessons, courses)
	ctx := context.Background()

	ch, err := s.CreateChapter(ctx, c.ID(), "Chuong 1")
	require.NoError(t, err)
	l, err := s.CreateLesson(ctx, ch.ID(), "Bai 1")
	require.NoError(t, err)

	return s, ch.ID(), l.ID()
}

func TestService_RenameChapter_OwnershipCheck(t *testing.T) {
	s, chapterID, _ := setupOwnership(t)
	ctx := context.Background()

	ch, err := s.RenameChapter(ctx, chapterID, "Chuong moi", 1, user.RoleTeacher)
	require.NoError(t, err, "giảng viên sở hữu phải sửa được")
	assert.Equal(t, "Chuong moi", ch.Title())

	_, err = s.RenameChapter(ctx, chapterID, "Hack", 999, user.RoleTeacher)
	assert.ErrorIs(t, err, curriculum.ErrChapterNotFound, "giảng viên khác không được sửa")

	_, err = s.RenameChapter(ctx, chapterID, "Admin sua", 999, user.RoleAdmin)
	assert.NoError(t, err, "admin luôn được phép")
}

func TestService_DeleteChapter_OwnershipAndNotEmpty(t *testing.T) {
	s, chapterID, _ := setupOwnership(t)
	ctx := context.Background()

	err := s.DeleteChapter(ctx, chapterID, 999, user.RoleTeacher)
	assert.ErrorIs(t, err, curriculum.ErrChapterNotFound, "giảng viên khác không được xóa")

	err = s.DeleteChapter(ctx, chapterID, 1, user.RoleTeacher)
	assert.NoError(t, err, "giảng viên sở hữu phải xóa được (fake repo không mô phỏng ràng buộc khóa ngoại)")
}

func TestService_RenameLesson_OwnershipCheck(t *testing.T) {
	s, _, lessonID := setupOwnership(t)
	ctx := context.Background()

	l, err := s.RenameLesson(ctx, lessonID, "Bai moi", 1, user.RoleTeacher)
	require.NoError(t, err)
	assert.Equal(t, "Bai moi", l.Title())

	_, err = s.RenameLesson(ctx, lessonID, "Hack", 999, user.RoleTeacher)
	assert.ErrorIs(t, err, curriculum.ErrLessonNotFound)
}

func TestService_DeleteLesson_OwnershipCheck(t *testing.T) {
	s, _, lessonID := setupOwnership(t)
	ctx := context.Background()

	err := s.DeleteLesson(ctx, lessonID, 999, user.RoleTeacher)
	assert.ErrorIs(t, err, curriculum.ErrLessonNotFound)

	err = s.DeleteLesson(ctx, lessonID, 1, user.RoleTeacher)
	assert.NoError(t, err)
}

// setupReorderFixture (US4.7) dựng course(teacherID=1) với 3 chương, mỗi
// chương có 2 bài học, để test sắp xếp lại thứ tự.
func setupReorderFixture(t *testing.T) (s *curriculumservice.Service, courseID uint, chapterIDs []uint, lessonIDs []uint) {
	t.Helper()
	c, err := course.NewCourse("Golang", "desc", 1)
	require.NoError(t, err)
	c.SetID(20)

	courses := newFakeCourseGetter()
	courses.byID[c.ID()] = c

	chapters := newFakeChapterRepo()
	lessons := newFakeLessonRepo()
	s = curriculumservice.NewService(chapters, lessons, courses)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		ch, err := s.CreateChapter(ctx, c.ID(), "Chuong")
		require.NoError(t, err)
		chapterIDs = append(chapterIDs, ch.ID())
	}
	for i := 0; i < 2; i++ {
		l, err := s.CreateLesson(ctx, chapterIDs[0], "Bai")
		require.NoError(t, err)
		lessonIDs = append(lessonIDs, l.ID())
	}
	return s, c.ID(), chapterIDs, lessonIDs
}

func TestService_ReorderChapters(t *testing.T) {
	s, courseID, chapterIDs, _ := setupReorderFixture(t)
	ctx := context.Background()

	// Đảo ngược thứ tự 3 chương.
	newOrder := []uint{chapterIDs[2], chapterIDs[1], chapterIDs[0]}
	reordered, err := s.ReorderChapters(ctx, courseID, newOrder, 1, user.RoleTeacher)
	require.NoError(t, err)
	require.Len(t, reordered, 3)
	for i, ch := range reordered {
		assert.Equal(t, newOrder[i], ch.ID())
		assert.Equal(t, i, ch.Position())
	}

	chapters, err := s.ListChapters(ctx, courseID)
	require.NoError(t, err)
	assert.Equal(t, chapterIDs[2], chapters[0].ID(), "chương đứng đầu sau khi sắp xếp lại phải đúng thứ tự mới")
}

func TestService_ReorderChapters_InvalidIDList(t *testing.T) {
	s, courseID, chapterIDs, _ := setupReorderFixture(t)
	ctx := context.Background()

	_, err := s.ReorderChapters(ctx, courseID, chapterIDs[:2], 1, user.RoleTeacher)
	assert.ErrorIs(t, err, curriculum.ErrInvalidChapterOrder, "thiếu 1 id phải bị từ chối")

	_, err = s.ReorderChapters(ctx, courseID, []uint{chapterIDs[0], chapterIDs[1], 9999}, 1, user.RoleTeacher)
	assert.ErrorIs(t, err, curriculum.ErrInvalidChapterOrder, "id không tồn tại phải bị từ chối")

	_, err = s.ReorderChapters(ctx, courseID, []uint{chapterIDs[0], chapterIDs[0], chapterIDs[1]}, 1, user.RoleTeacher)
	assert.ErrorIs(t, err, curriculum.ErrInvalidChapterOrder, "id lặp lại phải bị từ chối")
}

func TestService_ReorderChapters_OwnershipCheck(t *testing.T) {
	s, courseID, chapterIDs, _ := setupReorderFixture(t)
	ctx := context.Background()

	_, err := s.ReorderChapters(ctx, courseID, chapterIDs, 999, user.RoleTeacher)
	assert.ErrorIs(t, err, curriculum.ErrChapterNotFound, "giảng viên khác không được sắp xếp lại")

	_, err = s.ReorderChapters(ctx, courseID, chapterIDs, 999, user.RoleAdmin)
	assert.NoError(t, err, "admin luôn được phép")
}

func TestService_ReorderLessons(t *testing.T) {
	s, _, chapterIDs, lessonIDs := setupReorderFixture(t)
	ctx := context.Background()

	newOrder := []uint{lessonIDs[1], lessonIDs[0]}
	reordered, err := s.ReorderLessons(ctx, chapterIDs[0], newOrder, 1, user.RoleTeacher)
	require.NoError(t, err)
	require.Len(t, reordered, 2)
	assert.Equal(t, lessonIDs[1], reordered[0].ID())
	assert.Equal(t, 0, reordered[0].Position())
	assert.Equal(t, lessonIDs[0], reordered[1].ID())
	assert.Equal(t, 1, reordered[1].Position())
}

func TestService_ReorderLessons_InvalidIDList(t *testing.T) {
	s, _, chapterIDs, lessonIDs := setupReorderFixture(t)
	ctx := context.Background()

	_, err := s.ReorderLessons(ctx, chapterIDs[0], lessonIDs[:1], 1, user.RoleTeacher)
	assert.ErrorIs(t, err, curriculum.ErrInvalidLessonOrder)
}

func TestService_ReorderLessons_OwnershipCheck(t *testing.T) {
	s, _, chapterIDs, lessonIDs := setupReorderFixture(t)
	ctx := context.Background()

	_, err := s.ReorderLessons(ctx, chapterIDs[0], lessonIDs, 999, user.RoleTeacher)
	assert.ErrorIs(t, err, curriculum.ErrLessonNotFound)
}
