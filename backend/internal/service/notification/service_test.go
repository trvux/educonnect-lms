package notification_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/enrollment"
	"educonnect-lms/backend/internal/domain/notification"
	notificationservice "educonnect-lms/backend/internal/service/notification"

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

type fakeEnrollmentLister struct {
	items map[uint][]*enrollment.Enrollment
}

func (r *fakeEnrollmentLister) ListByCourse(_ context.Context, courseID uint) ([]*enrollment.Enrollment, error) {
	return r.items[courseID], nil
}

type fakeNotificationRepo struct {
	items  []*notification.Notification
	nextID uint
}

func (r *fakeNotificationRepo) CreateMany(_ context.Context, notifications []*notification.Notification) error {
	for _, n := range notifications {
		r.nextID++
		n.SetID(r.nextID)
		r.items = append(r.items, n)
	}
	return nil
}
func (r *fakeNotificationRepo) FindByID(_ context.Context, id uint) (*notification.Notification, error) {
	for _, n := range r.items {
		if n.ID() == id {
			return n, nil
		}
	}
	return nil, notification.ErrNotFound
}
func (r *fakeNotificationRepo) ListByRecipient(_ context.Context, recipientID uint) ([]*notification.Notification, error) {
	var out []*notification.Notification
	for _, n := range r.items {
		if n.RecipientID() == recipientID {
			out = append(out, n)
		}
	}
	return out, nil
}
func (r *fakeNotificationRepo) CountUnread(_ context.Context, recipientID uint) (int, error) {
	count := 0
	for _, n := range r.items {
		if n.RecipientID() == recipientID && !n.Read() {
			count++
		}
	}
	return count, nil
}
func (r *fakeNotificationRepo) Update(_ context.Context, updated *notification.Notification) error {
	for _, n := range r.items {
		if n.ID() == updated.ID() {
			return nil
		}
	}
	return notification.ErrNotFound
}

func newCourse(id uint) *course.Course {
	c, _ := course.NewCourse("Golang", "d", 1)
	c.SetID(id)
	return c
}

func TestService_SendToCourse(t *testing.T) {
	courses := &fakeCourseGetter{items: map[uint]*course.Course{1: newCourse(1)}}
	e1, _ := enrollment.NewEnrollment(10, 1)
	e2, _ := enrollment.NewEnrollment(20, 1)
	enrollments := &fakeEnrollmentLister{items: map[uint][]*enrollment.Enrollment{1: {e1, e2}}}
	notifications := &fakeNotificationRepo{}
	svc := notificationservice.NewService(notifications, enrollments, courses)

	sent, err := svc.SendToCourse(context.Background(), 1, "Bai tap moi", "noi dung")
	require.NoError(t, err)
	require.Len(t, sent, 2)

	list, err := svc.ListMine(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, list, 1)

	unread, err := svc.UnreadCount(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, 1, unread)
}

func TestService_SendToCourse_CourseNotFound(t *testing.T) {
	svc := notificationservice.NewService(&fakeNotificationRepo{}, &fakeEnrollmentLister{}, &fakeCourseGetter{items: map[uint]*course.Course{}})

	_, err := svc.SendToCourse(context.Background(), 999, "tieu de", "")
	assert.ErrorIs(t, err, course.ErrNotFound)
}

func TestService_MarkRead(t *testing.T) {
	courses := &fakeCourseGetter{items: map[uint]*course.Course{1: newCourse(1)}}
	e1, _ := enrollment.NewEnrollment(10, 1)
	enrollments := &fakeEnrollmentLister{items: map[uint][]*enrollment.Enrollment{1: {e1}}}
	notifications := &fakeNotificationRepo{}
	svc := notificationservice.NewService(notifications, enrollments, courses)

	sent, err := svc.SendToCourse(context.Background(), 1, "Bai tap moi", "")
	require.NoError(t, err)

	require.NoError(t, svc.MarkRead(context.Background(), sent[0].ID(), 10))

	unread, err := svc.UnreadCount(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, 0, unread)
}

func TestService_MarkRead_WrongRecipient(t *testing.T) {
	courses := &fakeCourseGetter{items: map[uint]*course.Course{1: newCourse(1)}}
	e1, _ := enrollment.NewEnrollment(10, 1)
	enrollments := &fakeEnrollmentLister{items: map[uint][]*enrollment.Enrollment{1: {e1}}}
	notifications := &fakeNotificationRepo{}
	svc := notificationservice.NewService(notifications, enrollments, courses)

	sent, err := svc.SendToCourse(context.Background(), 1, "Bai tap moi", "")
	require.NoError(t, err)

	err = svc.MarkRead(context.Background(), sent[0].ID(), 999)
	assert.ErrorIs(t, err, notification.ErrNotFound)
}
