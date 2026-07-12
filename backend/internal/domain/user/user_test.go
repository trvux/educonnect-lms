package user_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		fullName string
		role     user.Role
		wantErr  error
	}{
		{"valid student", "huy@vlu.edu.vn", "Huynh Bao Huy", user.RoleStudent, nil},
		{"invalid email", "not-an-email", "Huy", user.RoleStudent, user.ErrInvalidEmail},
		{"empty full name", "huy@vlu.edu.vn", "", user.RoleStudent, user.ErrEmptyFullName},
		{"invalid role", "huy@vlu.edu.vn", "Huy", user.Role("hacker"), user.ErrInvalidRole},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := user.NewUser(tt.email, tt.fullName, tt.role)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, u)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.email, u.Email())
			assert.True(t, u.Active(), "new user must be active by default")
		})
	}
}

func TestUser_SetPasswordHash(t *testing.T) {
	u, err := user.NewUser("huy@vlu.edu.vn", "Huy", user.RoleStudent)
	require.NoError(t, err)

	err = u.SetPasswordHash("")
	assert.ErrorIs(t, err, user.ErrEmptyPasswordHash)

	err = u.SetPasswordHash("$2a$10$hashedvalue")
	require.NoError(t, err)
	assert.Equal(t, "$2a$10$hashedvalue", u.PasswordHash())
}

func TestUser_DeactivateBlocksLogin(t *testing.T) {
	u, err := user.NewUser("huy@vlu.edu.vn", "Huy", user.RoleStudent)
	require.NoError(t, err)

	require.NoError(t, u.CanLogin())

	u.Deactivate()
	assert.ErrorIs(t, u.CanLogin(), user.ErrInactive)

	u.Activate()
	assert.NoError(t, u.CanLogin())
}
