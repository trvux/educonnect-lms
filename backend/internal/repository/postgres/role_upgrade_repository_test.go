package postgres_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/roleupgrade"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleUpgradeRepository(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	u, _ := user.NewUser("hv10@vlu.edu.vn", "Hoc Vien", user.RoleStudent)
	require.NoError(t, u.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, u))

	admin, _ := user.NewUser("admin10@vlu.edu.vn", "Admin", user.RoleAdmin)
	require.NoError(t, admin.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, admin))

	repo := postgres.NewRoleUpgradeRepository(pool)

	req, err := roleupgrade.NewRequest(u.ID(), "Em muốn dạy khóa Golang")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, req))
	assert.NotZero(t, req.ID())

	pending, err := repo.FindPendingByUser(ctx, u.ID())
	require.NoError(t, err)
	assert.Equal(t, req.ID(), pending.ID())

	list, err := repo.ListPending(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, req.Approve(admin.ID()))
	require.NoError(t, repo.Update(ctx, req))

	found, err := repo.FindByID(ctx, req.ID())
	require.NoError(t, err)
	assert.Equal(t, roleupgrade.StatusApproved, found.Status())
	require.NotNil(t, found.ReviewedBy())
	assert.Equal(t, admin.ID(), *found.ReviewedBy())

	_, err = repo.FindPendingByUser(ctx, u.ID())
	assert.ErrorIs(t, err, roleupgrade.ErrNotFound)
}
