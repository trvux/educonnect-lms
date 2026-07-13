package postgres_test

import (
	"context"
	"testing"
	"time"

	"educonnect-lms/backend/internal/domain/passwordreset"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOTPHasher struct{}

func (fakeOTPHasher) Compare(hash, plain string) error {
	if hash != "hash:"+plain {
		return assert.AnError
	}
	return nil
}

func TestPasswordResetRepository(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	userRepo := postgres.NewUserRepository(pool)
	u, _ := user.NewUser("huy10@vlu.edu.vn", "Huy", user.RoleStudent)
	require.NoError(t, u.SetPasswordHash("hash"))
	require.NoError(t, userRepo.Create(ctx, u))

	repo := postgres.NewPasswordResetRepository(pool)

	reset := passwordreset.New(u.ID(), "hash:123456", 10*time.Minute)
	require.NoError(t, repo.Create(ctx, reset))
	assert.NotZero(t, reset.ID())

	found, err := repo.FindActiveByUser(ctx, u.ID())
	require.NoError(t, err)
	assert.Equal(t, reset.ID(), found.ID())

	require.NoError(t, found.Verify("123456", fakeOTPHasher{}))
	require.NoError(t, repo.Update(ctx, found))

	reloaded, err := repo.FindActiveByUser(ctx, u.ID())
	require.NoError(t, err)
	assert.True(t, reloaded.Used())

	_, err = repo.FindActiveByUser(ctx, 999999)
	assert.ErrorIs(t, err, passwordreset.ErrNotFound)
}
