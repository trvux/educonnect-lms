package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestAuthHandler_Register_Success(t *testing.T) {
	svc := &fakeAuthService{
		registerFn: func(_ context.Context, in auth.RegisterInput) (*user.User, error) {
			u, err := user.NewUser(in.Email, in.FullName, in.Role)
			require.NoError(t, err)
			u.SetID(1)
			return u, nil
		},
	}
	h := handler.NewAuthHandler(svc, zap.NewNop())

	body, _ := json.Marshal(map[string]string{
		"email": "huy@vlu.edu.vn", "password": "secret123", "full_name": "Huy", "role": "student",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "huy@vlu.edu.vn")
}

func TestAuthHandler_Register_EmailTaken(t *testing.T) {
	svc := &fakeAuthService{
		registerFn: func(_ context.Context, _ auth.RegisterInput) (*user.User, error) {
			return nil, auth.ErrEmailTaken
		},
	}
	h := handler.NewAuthHandler(svc, zap.NewNop())

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
	h := handler.NewAuthHandler(svc, zap.NewNop())

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
