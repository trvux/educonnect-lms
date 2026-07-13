package postgres_test

import (
	"context"
	"testing"

	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndFindByEmail(t *testing.T) {
	pool := openTestPool(t)
	repo := postgres.NewUserRepository(pool)
	ctx := context.Background()

	u, err := user.NewUser("huy@vlu.edu.vn", "Huynh Bao Huy", user.RoleStudent)
	require.NoError(t, err)
	require.NoError(t, u.SetPasswordHash("hashed-value"))

	require.NoError(t, repo.Create(ctx, u))
	assert.NotZero(t, u.ID(), "Create must populate the generated id")

	found, err := repo.FindByEmail(ctx, "huy@vlu.edu.vn")
	require.NoError(t, err)
	assert.Equal(t, u.ID(), found.ID())
	assert.Equal(t, "Huynh Bao Huy", found.FullName())
	assert.Equal(t, user.RoleStudent, found.Role())
	assert.True(t, found.Active())
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	pool := openTestPool(t)
	repo := postgres.NewUserRepository(pool)

	_, err := repo.FindByEmail(context.Background(), "nobody@vlu.edu.vn")
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestUserRepository_UpdateProfileAndFindByPhone(t *testing.T) {
	pool := openTestPool(t)
	repo := postgres.NewUserRepository(pool)
	ctx := context.Background()

	u, err := user.NewUser("huy2@vlu.edu.vn", "Huy", user.RoleStudent)
	require.NoError(t, err)
	require.NoError(t, u.SetPasswordHash("hash"))
	require.NoError(t, repo.Create(ctx, u))

	require.NoError(t, u.UpdateProfile("Huynh Bao Huy", "0987654321", "2074802010001"))
	u.SetAvatarPath("avatars/1/photo.jpg")
	require.NoError(t, repo.Update(ctx, u))

	found, err := repo.FindByPhone(ctx, "0987654321")
	require.NoError(t, err)
	assert.Equal(t, u.ID(), found.ID())
	assert.Equal(t, "Huynh Bao Huy", found.FullName())
	assert.Equal(t, "2074802010001", found.StudentCode())
	assert.Equal(t, "avatars/1/photo.jpg", found.AvatarPath())

	_, err = repo.FindByPhone(ctx, "0900000000")
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestUserRepository_Update(t *testing.T) {
	pool := openTestPool(t)
	repo := postgres.NewUserRepository(pool)
	ctx := context.Background()

	u, err := user.NewUser("huy@vlu.edu.vn", "Huy", user.RoleStudent)
	require.NoError(t, err)
	require.NoError(t, u.SetPasswordHash("hash"))
	require.NoError(t, repo.Create(ctx, u))

	u.Deactivate() // US1.3
	require.NoError(t, repo.Update(ctx, u))

	found, err := repo.FindByID(ctx, u.ID())
	require.NoError(t, err)
	assert.False(t, found.Active())
}
