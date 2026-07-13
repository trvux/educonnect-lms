package roleupgrade_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/roleupgrade"
	"educonnect-lms/backend/internal/domain/user"
	roleupgradeservice "educonnect-lms/backend/internal/service/roleupgrade"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRequestRepo struct {
	items  []*roleupgrade.Request
	nextID uint
}

func (r *fakeRequestRepo) Create(_ context.Context, req *roleupgrade.Request) error {
	r.nextID++
	req.SetID(r.nextID)
	r.items = append(r.items, req)
	return nil
}
func (r *fakeRequestRepo) FindByID(_ context.Context, id uint) (*roleupgrade.Request, error) {
	for _, req := range r.items {
		if req.ID() == id {
			return req, nil
		}
	}
	return nil, roleupgrade.ErrNotFound
}
func (r *fakeRequestRepo) FindPendingByUser(_ context.Context, userID uint) (*roleupgrade.Request, error) {
	for _, req := range r.items {
		if req.UserID() == userID && req.Status() == roleupgrade.StatusPending {
			return req, nil
		}
	}
	return nil, roleupgrade.ErrNotFound
}
func (r *fakeRequestRepo) ListPending(_ context.Context) ([]*roleupgrade.Request, error) {
	var out []*roleupgrade.Request
	for _, req := range r.items {
		if req.Status() == roleupgrade.StatusPending {
			out = append(out, req)
		}
	}
	return out, nil
}
func (r *fakeRequestRepo) Update(_ context.Context, req *roleupgrade.Request) error { return nil }

type fakeUserRepo struct{ items map[uint]*user.User }

func (r *fakeUserRepo) Create(_ context.Context, u *user.User) error { return nil }
func (r *fakeUserRepo) FindByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrNotFound
}
func (r *fakeUserRepo) FindByID(_ context.Context, id uint) (*user.User, error) {
	if u, ok := r.items[id]; ok {
		return u, nil
	}
	return nil, user.ErrNotFound
}
func (r *fakeUserRepo) FindByPhone(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrNotFound
}
func (r *fakeUserRepo) Update(_ context.Context, u *user.User) error { return nil }

func newStudent(id uint) *user.User {
	u, _ := user.NewUser("hv@vlu.edu.vn", "Hoc Vien", user.RoleStudent)
	u.SetID(id)
	return u
}

func TestService_CreateAndApprove(t *testing.T) {
	requests := &fakeRequestRepo{}
	users := &fakeUserRepo{items: map[uint]*user.User{1: newStudent(1)}}
	svc := roleupgradeservice.NewService(requests, users)
	ctx := context.Background()

	req, err := svc.Create(ctx, 1, "Em muốn dạy khóa Golang")
	require.NoError(t, err)

	list, err := svc.ListPending(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)

	approved, err := svc.Approve(ctx, req.ID(), 99)
	require.NoError(t, err)
	assert.Equal(t, roleupgrade.StatusApproved, approved.Status())

	promoted, err := users.FindByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, user.RoleTeacher, promoted.Role())
}

func TestService_Create_AlreadyPending(t *testing.T) {
	requests := &fakeRequestRepo{}
	users := &fakeUserRepo{items: map[uint]*user.User{1: newStudent(1)}}
	svc := roleupgradeservice.NewService(requests, users)
	ctx := context.Background()

	_, err := svc.Create(ctx, 1, "ly do 1")
	require.NoError(t, err)

	_, err = svc.Create(ctx, 1, "ly do 2")
	assert.ErrorIs(t, err, roleupgrade.ErrAlreadyPending)
}

func TestService_Reject(t *testing.T) {
	requests := &fakeRequestRepo{}
	users := &fakeUserRepo{items: map[uint]*user.User{1: newStudent(1)}}
	svc := roleupgradeservice.NewService(requests, users)
	ctx := context.Background()

	req, err := svc.Create(ctx, 1, "ly do")
	require.NoError(t, err)

	rejected, err := svc.Reject(ctx, req.ID(), 99)
	require.NoError(t, err)
	assert.Equal(t, roleupgrade.StatusRejected, rejected.Status())

	stillStudent, err := users.FindByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, user.RoleStudent, stillStudent.Role())
}
