package enrollment_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/enrollment"
	"educonnect-lms/backend/internal/domain/user"
	enrollmentservice "educonnect-lms/backend/internal/service/enrollment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeEnrollmentRepo struct {
	items  []*enrollment.Enrollment
	nextID uint
}

func (r *fakeEnrollmentRepo) Create(_ context.Context, e *enrollment.Enrollment) error {
	r.nextID++
	e.SetID(r.nextID)
	r.items = append(r.items, e)
	return nil
}
func (r *fakeEnrollmentRepo) IsEnrolled(_ context.Context, studentID, courseID uint) (bool, error) {
	for _, e := range r.items {
		if e.StudentID() == studentID && e.CourseID() == courseID {
			return true, nil
		}
	}
	return false, nil
}
func (r *fakeEnrollmentRepo) ListByCourse(_ context.Context, courseID uint) ([]*enrollment.Enrollment, error) {
	var out []*enrollment.Enrollment
	for _, e := range r.items {
		if e.CourseID() == courseID {
			out = append(out, e)
		}
	}
	return out, nil
}
func (r *fakeEnrollmentRepo) ListByStudent(_ context.Context, studentID uint) ([]*enrollment.Enrollment, error) {
	var out []*enrollment.Enrollment
	for _, e := range r.items {
		if e.StudentID() == studentID {
			out = append(out, e)
		}
	}
	return out, nil
}

type fakeCourseRepo struct{ byID map[uint]*course.Course }

func (r *fakeCourseRepo) Create(_ context.Context, c *course.Course) error { return nil }
func (r *fakeCourseRepo) FindByID(_ context.Context, id uint) (*course.Course, error) {
	if c, ok := r.byID[id]; ok {
		return c, nil
	}
	return nil, course.ErrNotFound
}
func (r *fakeCourseRepo) Search(_ context.Context, _ string) ([]*course.Course, error) {
	return nil, nil
}
func (r *fakeCourseRepo) ListByTeacher(_ context.Context, _ uint) ([]*course.Course, error) {
	return nil, nil
}
func (r *fakeCourseRepo) ListByStatus(_ context.Context, _ course.Status) ([]*course.Course, error) {
	return nil, nil
}
func (r *fakeCourseRepo) Update(_ context.Context, c *course.Course) error {
	r.byID[c.ID()] = c
	return nil
}

type fakeUserRepo struct{ byID map[uint]*user.User }

func (r *fakeUserRepo) Create(_ context.Context, u *user.User) error { return nil }
func (r *fakeUserRepo) FindByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrNotFound
}
func (r *fakeUserRepo) FindByID(_ context.Context, id uint) (*user.User, error) {
	if u, ok := r.byID[id]; ok {
		return u, nil
	}
	return nil, user.ErrNotFound
}
func (r *fakeUserRepo) Update(_ context.Context, u *user.User) error { return nil }

func setup(t *testing.T) (*enrollmentservice.Service, *course.Course, *user.User) {
	t.Helper()
	teacher, err := user.NewUser("gv@vlu.edu.vn", "GV", user.RoleTeacher)
	require.NoError(t, err)
	teacher.SetID(1)

	student, err := user.NewUser("hv@vlu.edu.vn", "Hoc Vien A", user.RoleStudent)
	require.NoError(t, err)
	student.SetID(2)

	c, err := course.NewCourse("Golang", "desc", teacher.ID())
	require.NoError(t, err)
	c.SetID(1)
	c.SubmitForReview()
	require.NoError(t, c.Approve())

	courseRepo := &fakeCourseRepo{byID: map[uint]*course.Course{c.ID(): c}}
	userRepo := &fakeUserRepo{byID: map[uint]*user.User{teacher.ID(): teacher, student.ID(): student}}
	svc := enrollmentservice.NewService(&fakeEnrollmentRepo{}, courseRepo, userRepo)
	return svc, c, student
}

func TestService_Enroll(t *testing.T) {
	svc, c, student := setup(t)
	ctx := context.Background()

	e, err := svc.Enroll(ctx, student.ID(), c.ID())
	require.NoError(t, err)
	assert.Equal(t, student.ID(), e.StudentID())

	// Đăng ký lần 2 phải báo lỗi trùng.
	_, err = svc.Enroll(ctx, student.ID(), c.ID())
	assert.ErrorIs(t, err, enrollment.ErrAlreadyEnrolled)
}

func TestService_Enroll_CourseNotApproved(t *testing.T) {
	teacher, _ := user.NewUser("gv@vlu.edu.vn", "GV", user.RoleTeacher)
	teacher.SetID(1)
	student, _ := user.NewUser("hv@vlu.edu.vn", "HV", user.RoleStudent)
	student.SetID(2)
	draft, _ := course.NewCourse("Con Draft", "desc", teacher.ID())
	draft.SetID(1)

	courseRepo := &fakeCourseRepo{byID: map[uint]*course.Course{draft.ID(): draft}}
	userRepo := &fakeUserRepo{byID: map[uint]*user.User{}}
	svc := enrollmentservice.NewService(&fakeEnrollmentRepo{}, courseRepo, userRepo)

	_, err := svc.Enroll(context.Background(), student.ID(), draft.ID())
	assert.ErrorIs(t, err, enrollmentservice.ErrCourseNotApproved)
}

func TestService_ListStudents_OnlyOwningTeacher(t *testing.T) {
	svc, c, student := setup(t)
	ctx := context.Background()

	_, err := svc.Enroll(ctx, student.ID(), c.ID())
	require.NoError(t, err)

	list, err := svc.ListStudents(ctx, c.ID(), c.TeacherID())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "Hoc Vien A", list[0].FullName)

	_, err = svc.ListStudents(ctx, c.ID(), 999) // giáo viên khác
	assert.ErrorIs(t, err, course.ErrNotFound)
}
