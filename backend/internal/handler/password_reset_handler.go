package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"educonnect-lms/backend/internal/domain/passwordreset"

	"go.uber.org/zap"
)

// PasswordResetService là tập con method của *auth.Service mà handler cần
// (US1.6).
type PasswordResetService interface {
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, otp, newPassword string) error
}

type PasswordResetHandler struct {
	service PasswordResetService
	log     *zap.Logger
}

func NewPasswordResetHandler(service PasswordResetService, log *zap.Logger) *PasswordResetHandler {
	return &PasswordResetHandler{service: service, log: log}
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

// Forgot xử lý POST /api/auth/forgot-password (US1.6, public). Luôn trả
// 200 dù email không tồn tại — không tiết lộ tài khoản (chống dò email).
func (h *PasswordResetHandler) Forgot(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}
	if err := h.service.ForgotPassword(r.Context(), req.Email); err != nil {
		h.log.Error("password reset handler: gửi OTP thất bại", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Nếu email tồn tại trong hệ thống, mã OTP đã được gửi.",
	})
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

// Reset xử lý POST /api/auth/reset-password (US1.6, public).
func (h *PasswordResetHandler) Reset(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}
	err := h.service.ResetPassword(r.Context(), req.Email, req.OTP, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, passwordreset.ErrInvalidOTP):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, passwordreset.ErrTooManyAttempts):
			writeError(w, http.StatusTooManyRequests, err.Error())
		default:
			h.log.Error("password reset handler: đặt lại mật khẩu thất bại", zap.Error(err))
			writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
		}
		return
	}
	writeJSON(w, http.StatusOK, nil)
}
