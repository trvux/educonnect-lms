package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/handler"
	"educonnect-lms/backend/internal/service/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeAuthService struct {
	registerFn func(ctx context.Context, in auth.RegisterInput) (*user.User, error)
	loginFn    func(ctx context.Context, in auth.LoginInput) (string, error)
}

func (f *fakeAuthService) Register(ctx context.Context, in auth.RegisterInput) (*user.User, error) {
	return f.registerFn(ctx, in)
}

func (f *fakeAuthService) Login(ctx context.Context, in auth.LoginInput) (string, error) {
	return f.loginFn(ctx, in)
}

func (f *fakeAuthService) GetProfile(_ context.Context, _ uint) (*user.User, error) {
	return nil, user.ErrNotFound
}

func (f *fakeAuthService) UpdateProfile(_ context.Context, _ uint, _, _, _ string) (*user.User, error) {
	return nil, user.ErrNotFound
}

func (f *fakeAuthService) ChangePassword(_ context.Context, _ uint, _, _ string) error {
	return auth.ErrInvalidCredentials
}

func (f *fakeAuthService) UploadAvatar(_ context.Context, _ uint, _ string, _ io.Reader) (*user.User, error) {
	return nil, user.ErrNotFound
}

func (f *fakeAuthService) ForgotUsername(_ context.Context, _ string) (string, error) {
	return "", user.ErrNotFound
}

func (f *fakeAuthService) VerifyEmail(_ context.Context, _, _ string) error {
	return nil
}

func (f *fakeAuthService) ResendVerification(_ context.Context, _ string) error {
	return nil
}

func TestAuthHandler_Register_Success(t *testing.T) {
	svc := &fakeAuthService{
		registerFn: func(_ context.Context, in auth.RegisterInput) (*user.User, error) {
			u, err := user.NewUser(in.Email, in.FullName, in.Role)
			require.NoError(t, err)
			u.SetID(1)
			return u, nil
		},
	}
	h := handler.NewAuthHandler(svc, zap.NewNop(), false)

	body, _ := json.Marshal(map[string]string{
		"email": "huy@vlu.edu.vn", "password": "secret123", "full_name": "Huy", "role": "student",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "huy@vlu.edu.vn")
}

func TestAuthHandler_Register_ForcesStudent_WhenRoleOverrideDisabled(t *testing.T) {
	var gotRole user.Role
	var gotSkip bool
	svc := &fakeAuthService{
		registerFn: func(_ context.Context, in auth.RegisterInput) (*user.User, error) {
			gotRole = in.Role
			gotSkip = in.SkipEmailVerification
			u, err := user.NewUser(in.Email, in.FullName, in.Role)
			require.NoError(t, err)
			u.SetID(1)
			return u, nil
		},
	}
	h := handler.NewAuthHandler(svc, zap.NewNop(), false) // ALLOW_ROLE_ON_REGISTER=false (US1.7)

	body, _ := json.Marshal(map[string]string{
		"email": "huy@vlu.edu.vn", "password": "secret123", "full_name": "Huy", "role": "teacher",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, user.RoleStudent, gotRole, "role client gửi lên phải bị bỏ qua khi flag tắt")
	assert.False(t, gotSkip, "production phải bắt buộc xác thực email (US1.9)")
}

func TestAuthHandler_Register_HonorsRole_WhenOverrideEnabled(t *testing.T) {
	var gotRole user.Role
	var gotSkip bool
	svc := &fakeAuthService{
		registerFn: func(_ context.Context, in auth.RegisterInput) (*user.User, error) {
			gotRole = in.Role
			gotSkip = in.SkipEmailVerification
			u, err := user.NewUser(in.Email, in.FullName, in.Role)
			require.NoError(t, err)
			u.SetID(1)
			return u, nil
		},
	}
	h := handler.NewAuthHandler(svc, zap.NewNop(), true) // dev/seed only

	body, _ := json.Marshal(map[string]string{
		"email": "gv@vlu.edu.vn", "password": "secret123", "full_name": "GV", "role": "teacher",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, user.RoleTeacher, gotRole)
	assert.True(t, gotSkip, "dev/seed phải bỏ qua OTP xác thực email")
}

func TestAuthHandler_Register_EmailTaken(t *testing.T) {
	svc := &fakeAuthService{
		registerFn: func(_ context.Context, _ auth.RegisterInput) (*user.User, error) {
			return nil, auth.ErrEmailTaken
		},
	}
	h := handler.NewAuthHandler(svc, zap.NewNop(), false)

	body, _ := json.Marshal(map[string]string{
		"email": "huy@vlu.edu.vn", "password": "secret123", "full_name": "Huy", "role": "student",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAuthHandler_Login(t *testing.T) {
	svc := &fakeAuthService{
		loginFn: func(_ context.Context, in auth.LoginInput) (string, error) {
			if in.Password != "secret123" {
				return "", auth.ErrInvalidCredentials
			}
			return "signed-jwt-token", nil
		},
	}
	h := handler.NewAuthHandler(svc, zap.NewNop(), false)

	t.Run("valid credentials", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "huy@vlu.edu.vn", "password": "secret123"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "signed-jwt-token")
	})

	t.Run("wrong password", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "huy@vlu.edu.vn", "password": "wrong"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
