package auth_test

import (
	"context"
	"errors"
	"testing"

	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/service/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUserRepo is an in-memory stand-in for user.Repository so the service
// layer can be unit tested without a real Postgres instance.
type fakeUserRepo struct {
	byEmail map[string]*user.User
	nextID  uint
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]*user.User{}}
}

func (r *fakeUserRepo) Create(_ context.Context, u *user.User) error {
	r.nextID++
	u.SetID(r.nextID)
	r.byEmail[u.Email()] = u
	return nil
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if u, ok := r.byEmail[email]; ok {
		return u, nil
	}
	return nil, user.ErrNotFound
}

func (r *fakeUserRepo) FindByID(_ context.Context, id uint) (*user.User, error) {
	for _, u := range r.byEmail {
		if u.ID() == id {
			return u, nil
		}
	}
	return nil, user.ErrNotFound
}

func (r *fakeUserRepo) Update(_ context.Context, u *user.User) error {
	r.byEmail[u.Email()] = u
	return nil
}

// fakeHasher avoids pulling real bcrypt into the unit test (kept fast + pure).
type fakeHasher struct{}

func (fakeHasher) Hash(plain string) (string, error) { return "hashed:" + plain, nil }
func (fakeHasher) Compare(hash, plain string) error {
	if hash != "hashed:"+plain {
		return errors.New("mismatch")
	}
	return nil
}

type fakeTokenIssuer struct{}

func (fakeTokenIssuer) Issue(userID uint, role user.Role) (string, error) {
	return "token-for-user", nil
}

func newService() *auth.Service {
	return auth.NewService(newFakeUserRepo(), fakeHasher{}, fakeTokenIssuer{})
}

func TestService_Register(t *testing.T) {
	s := newService()
	ctx := context.Background()

	u, err := s.Register(ctx, auth.RegisterInput{
		Email:    "huy@vlu.edu.vn",
		Password: "secret123",
		FullName: "Huynh Bao Huy",
		Role:     user.RoleStudent,
	})
	require.NoError(t, err)
	assert.Equal(t, "huy@vlu.edu.vn", u.Email())
	assert.NotEmpty(t, u.PasswordHash())

	// Registering the same email twice must fail (US1.1 uniqueness rule).
	_, err = s.Register(ctx, auth.RegisterInput{
		Email:    "huy@vlu.edu.vn",
		Password: "other",
		FullName: "Duplicate",
		Role:     user.RoleStudent,
	})
	assert.ErrorIs(t, err, auth.ErrEmailTaken)
}

func TestService_Login(t *testing.T) {
	s := newService()
	ctx := context.Background()

	_, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
	})
	require.NoError(t, err)

	t.Run("correct credentials", func(t *testing.T) {
		token, err := s.Login(ctx, auth.LoginInput{Email: "huy@vlu.edu.vn", Password: "secret123"})
		require.NoError(t, err)
		assert.Equal(t, "token-for-user", token)
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := s.Login(ctx, auth.LoginInput{Email: "huy@vlu.edu.vn", Password: "wrong"})
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	})

	t.Run("unknown email", func(t *testing.T) {
		_, err := s.Login(ctx, auth.LoginInput{Email: "nobody@vlu.edu.vn", Password: "x"})
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	})
}
