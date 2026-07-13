package lessoncompletion_test

import (
	"context"
	"sort"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/lessoncompletion"
	"educonnect-lms/backend/internal/domain/user"
	lessoncompletionservice "educonnect-lms/backend/internal/service/lessoncompletion"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeChapterRepo struct{ byID map[uint]*curriculum.Chapter }

func newFakeChapterRepo() *fakeChapterRepo {
	return &fakeChapterRepo{byID: map[uint]*curriculum.Chapter{}}
}

func (r *fakeChapterRepo) Create(_ context.Context, c *curriculum.Chapter) error {
	c.SetID(uint(len(r.byID) + 1))
	r.byID[c.ID()] = c
	return nil
}
func (r *fakeChapterRepo) FindByID(_ context.Context, id uint) (*curriculum.Chapter, error) {
	c, ok := r.byID[id]
	if !ok {
		return nil, curriculum.ErrChapterNotFound
	}
	return c, nil
}
func (r *fakeChapterRepo) ListByCourse(_ context.Context, courseID uint) ([]*curriculum.Chapter, error) {
	var out []*curriculum.Chapter
	for _, c := range r.byID {
		if c.CourseID() == courseID {
			out = append(out, c)
		}
	}
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
	delete(r.byID, id)
	return nil
}

type fakeLessonRepo struct{ byID map[uint]*curriculum.Lesson }

func newFakeLessonRepo() *fakeLessonRepo { return &fakeLessonRepo{byID: map[uint]*curriculum.Lesson{}} }

func (r *fakeLessonRepo) Create(_ context.Context, l *curriculum.Lesson) error {
	l.SetID(uint(len(r.byID) + 1))
	r.byID[l.ID()] = l
	return nil
}
func (r *fakeLessonRepo) FindByID(_ context.Context, id uint) (*curriculum.Lesson, error) {
	l, ok := r.byID[id]
	if !ok {
		return nil, curriculum.ErrLessonNotFound
	}
	return l, nil
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
	delete(r.byID, id)
	return nil
}

type fakeCourseGetter struct{ byID map[uint]*course.Course }

func newFakeCourseGetter() *fakeCourseGetter {
	return &fakeCourseGetter{byID: map[uint]*course.Course{}}
}

func (r *fakeCourseGetter) FindByID(_ context.Context, id uint) (*course.Course, error) {
	c, ok := r.byID[id]
	if !ok {
		return nil, course.ErrNotFound
	}
	return c, nil
}

type fakeEnrollmentChecker struct{ enrolled map[uint]map[uint]bool } // studentID -> courseID -> true

func newFakeEnrollmentChecker() *fakeEnrollmentChecker {
	return &fakeEnrollmentChecker{enrolled: map[uint]map[uint]bool{}}
}
func (r *fakeEnrollmentChecker) enroll(studentID, courseID uint) {
	if r.enrolled[studentID] == nil {
		r.enrolled[studentID] = map[uint]bool{}
	}
	r.enrolled[studentID][courseID] = true
}
func (r *fakeEnrollmentChecker) IsEnrolled(_ context.Context, studentID, courseID uint) (bool, error) {
	return r.enrolled[studentID][courseID], nil
}

type fakeCompletionRepo struct {
	completed map[uint]map[uint]bool // studentID -> lessonID -> true
}

func newFakeCompletionRepo() *fakeCompletionRepo {
	return &fakeCompletionRepo{completed: map[uint]map[uint]bool{}}
}
func (r *fakeCompletionRepo) Create(_ context.Context, c *lessoncompletion.LessonCompletion) error {
	if r.completed[c.StudentID()] == nil {
		r.completed[c.StudentID()] = map[uint]bool{}
	}
	r.completed[c.StudentID()][c.LessonID()] = true
	return nil
}
func (r *fakeCompletionRepo) IsCompleted(_ context.Context, studentID, lessonID uint) (bool, error) {
	return r.completed[studentID][lessonID], nil
}
func (r *fakeCompletionRepo) ListCompletedByStudent(_ context.Context, studentID uint) (map[uint]bool, error) {
	out := map[uint]bool{}
	for id := range r.completed[studentID] {
		out[id] = true
	}
	return out, nil
}

// setupFixture dựng course(teacherID=1) với 2 chương, mỗi chương 2 bài học
// (tổng 4 bài, thứ tự: L1, L2 ở Chuong 1; L3, L4 ở Chuong 2), học viên id=2
// đã đăng ký khóa học.
func setupFixture(t *testing.T) (s *lessoncompletionservice.Service, courseID uint, lessonIDs []uint, enroll *fakeEnrollmentChecker) {
	t.Helper()
	c, err := course.NewCourse("Golang", "desc", 1)
	require.NoError(t, err)
	c.SetID(1)

	courses := newFakeCourseGetter()
	courses.byID[c.ID()] = c

	chapters := newFakeChapterRepo()
	lessons := newFakeLessonRepo()
	completions := newFakeCompletionRepo()
	enroll = newFakeEnrollmentChecker()
	enroll.enroll(2, c.ID())

	s = lessoncompletionservice.NewService(completions, lessons, chapters, courses, enroll)

	ctx := context.Background()
	for chIdx := 0; chIdx < 2; chIdx++ {
		ch, err := curriculum.NewChapter(c.ID(), "Chuong", chIdx)
		require.NoError(t, err)
		require.NoError(t, chapters.Create(ctx, ch))
		for lIdx := 0; lIdx < 2; lIdx++ {
			l, err := curriculum.NewLesson(ch.ID(), "Bai", lIdx)
			require.NoError(t, err)
			require.NoError(t, lessons.Create(ctx, l))
			lessonIDs = append(lessonIDs, l.ID())
		}
	}
	return s, c.ID(), lessonIDs, enroll
}

func TestService_ListForStudent_SequentialLock(t *testing.T) {
	s, courseID, lessonIDs, _ := setupFixture(t)
	ctx := context.Background()

	states, err := s.ListForStudent(ctx, courseID, 2, user.RoleStudent)
	require.NoError(t, err)
	require.Len(t, states, 4)

	// Chưa hoàn thành gì: chỉ bài đầu tiên mở khóa, còn lại đều khóa.
	assert.False(t, states[0].Locked, "bài đầu tiên luôn mở khóa")
	assert.False(t, states[0].Completed)
	for i := 1; i < 4; i++ {
		assert.True(t, states[i].Locked, "bài thứ %d phải bị khóa khi bài trước chưa hoàn thành", i+1)
	}

	// Hoàn thành bài 1 -> bài 2 mở khóa, bài 3/4 vẫn khóa.
	require.NoError(t, s.MarkComplete(ctx, lessonIDs[0], 2))
	states, err = s.ListForStudent(ctx, courseID, 2, user.RoleStudent)
	require.NoError(t, err)
	assert.True(t, states[0].Completed)
	assert.False(t, states[1].Locked, "bài 2 phải mở khóa sau khi hoàn thành bài 1")
	assert.True(t, states[2].Locked)
	assert.True(t, states[3].Locked)
}

func TestService_MarkComplete_RejectsLockedLesson(t *testing.T) {
	s, _, lessonIDs, _ := setupFixture(t)
	ctx := context.Background()

	// Cố hoàn thành bài 2 khi bài 1 chưa hoàn thành -> phải bị chặn.
	err := s.MarkComplete(ctx, lessonIDs[1], 2)
	assert.ErrorIs(t, err, lessoncompletion.ErrLessonLocked)
}

func TestService_MarkComplete_Idempotent(t *testing.T) {
	s, _, lessonIDs, _ := setupFixture(t)
	ctx := context.Background()

	require.NoError(t, s.MarkComplete(ctx, lessonIDs[0], 2))
	// Gọi lại lần 2 trên bài đã hoàn thành không được lỗi.
	assert.NoError(t, s.MarkComplete(ctx, lessonIDs[0], 2))
}

func TestService_MarkComplete_RequiresEnrollment(t *testing.T) {
	s, _, lessonIDs, _ := setupFixture(t)
	ctx := context.Background()

	// Học viên id=999 chưa đăng ký khóa học này.
	err := s.MarkComplete(ctx, lessonIDs[0], 999)
	assert.ErrorIs(t, err, curriculum.ErrLessonNotFound)
}

func TestService_ListForStudent_TeacherAlwaysUnlocked(t *testing.T) {
	s, courseID, _, _ := setupFixture(t)
	ctx := context.Background()

	states, err := s.ListForStudent(ctx, courseID, 1, user.RoleTeacher)
	require.NoError(t, err)
	for _, st := range states {
		assert.False(t, st.Locked, "giảng viên phải luôn thấy mọi bài mở khóa")
	}
}
