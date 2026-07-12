package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/service/auth"

	"go.uber.org/zap"
)

// AuthService is the subset of *auth.Service the handler depends on — kept
// as an interface so tests can inject a fake instead of hitting Postgres.
type AuthService interface {
	Register(ctx context.Context, in auth.RegisterInput) (*user.User, error)
	Login(ctx context.Context, in auth.LoginInput) (string, error)
}

type AuthHandler struct {
	service AuthService
	log     *zap.Logger
}

func NewAuthHandler(service AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{service: service, log: log}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type userResponse struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

// Register handles POST /api/auth/register (US1.1).
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	u, err := h.service.Register(r.Context(), auth.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     user.Role(req.Role),
	})
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, userResponse{
		ID: u.ID(), Email: u.Email(), FullName: u.FullName(), Role: string(u.Role()),
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Login handles POST /api/auth/login (US1.2).
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	token, err := h.service.Login(r.Context(), auth.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		h.handleAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, user.ErrInactive):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrEmptyFullName),
		errors.Is(err, user.ErrInvalidRole),
		errors.Is(err, user.ErrEmptyPasswordHash):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("auth handler: unexpected error", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
