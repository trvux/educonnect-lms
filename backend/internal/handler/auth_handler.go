package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/handler/middleware"
	"educonnect-lms/backend/internal/service/auth"

	"go.uber.org/zap"
)

const maxAvatarSize = 5 << 20 // 5MB, đủ cho ảnh đại diện

// AuthService là tập con method của *auth.Service mà handler cần — để dạng
// interface để test có thể inject fake thay vì gọi Postgres thật.
type AuthService interface {
	Register(ctx context.Context, in auth.RegisterInput) (*user.User, error)
	Login(ctx context.Context, in auth.LoginInput) (string, error)
	GetProfile(ctx context.Context, userID uint) (*user.User, error)
	UpdateProfile(ctx context.Context, userID uint, fullName, phone, studentCode string) (*user.User, error)
	ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error
	UploadAvatar(ctx context.Context, userID uint, fileName string, content io.Reader) (*user.User, error)
	ForgotUsername(ctx context.Context, phone string) (string, error)
}

type AuthHandler struct {
	service AuthService
	log     *zap.Logger
	// allowRoleOnRegister — xem config.AllowRoleOnRegister (US1.7): chỉ bật
	// ở dev/seed, production luôn ép Student bất kể client gửi gì.
	allowRoleOnRegister bool
}

func NewAuthHandler(service AuthService, log *zap.Logger, allowRoleOnRegister bool) *AuthHandler {
	return &AuthHandler{service: service, log: log, allowRoleOnRegister: allowRoleOnRegister}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	// Role chỉ được tôn trọng khi ALLOW_ROLE_ON_REGISTER=true (dev/seed).
	// Mặc định US1.7: đăng ký công khai luôn tạo Student.
	Role string `json:"role"`
}

type userResponse struct {
	ID          uint   `json:"id"`
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	Role        string `json:"role"`
	Phone       string `json:"phone,omitempty"`
	StudentCode string `json:"student_code,omitempty"`
	AvatarPath  string `json:"avatar_path,omitempty"`
}

func toUserResponse(u *user.User) userResponse {
	return userResponse{
		ID:          u.ID(),
		Email:       u.Email(),
		FullName:    u.FullName(),
		Role:        string(u.Role()),
		Phone:       u.Phone(),
		StudentCode: u.StudentCode(),
		AvatarPath:  u.AvatarPath(),
	}
}

// Register xử lý POST /api/auth/register (US1.1, sửa theo US1.7).
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON không hợp lệ")
		return
	}

	role := user.RoleStudent
	if h.allowRoleOnRegister && req.Role != "" {
		role = user.Role(req.Role)
	}

	u, err := h.service.Register(r.Context(), auth.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     role,
	})
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(u))
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Login xử lý POST /api/auth/login (US1.2).
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON không hợp lệ")
		return
	}

	token, err := h.service.Login(r.Context(), auth.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		h.handleAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

// Me xử lý GET /api/me (US1.4).
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	u, err := h.service.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

type updateProfileRequest struct {
	FullName    string `json:"full_name"`
	Phone       string `json:"phone"`
	StudentCode string `json:"student_code"`
}

// UpdateMe xử lý PATCH /api/me (US1.4).
func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}
	u, err := h.service.UpdateProfile(r.Context(), claims.UserID, req.FullName, req.Phone, req.StudentCode)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

// UploadAvatar xử lý POST /api/me/avatar (US1.4, multipart field "file").
func (h *AuthHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		writeError(w, http.StatusBadRequest, "file tải lên quá lớn hoặc không đúng định dạng multipart")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "thiếu file trong field \"file\"")
		return
	}
	defer file.Close()

	u, err := h.service.UploadAvatar(r.Context(), claims.UserID, header.Filename, file)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword xử lý POST /api/auth/change-password (US1.5).
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}
	if err := h.service.ChangePassword(r.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		h.handleAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

type forgotUsernameRequest struct {
	Phone string `json:"phone"`
}

type forgotUsernameResponse struct {
	MaskedEmail string `json:"masked_email"`
}

// ForgotUsername xử lý POST /api/auth/forgot-username (US1.8, public).
func (h *AuthHandler) ForgotUsername(w http.ResponseWriter, r *http.Request) {
	var req forgotUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}
	masked, err := h.service.ForgotUsername(r.Context(), req.Phone)
	if err != nil {
		// Không tiết lộ SĐT có tồn tại hay không — luôn trả 1 thông báo chung.
		if errors.Is(err, user.ErrNotFound) {
			writeError(w, http.StatusNotFound, "không tìm thấy tài khoản với số điện thoại này")
			return
		}
		h.handleAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, forgotUsernameResponse{MaskedEmail: masked})
}

func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, user.ErrInactive):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, user.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrEmptyFullName),
		errors.Is(err, user.ErrInvalidRole),
		errors.Is(err, user.ErrInvalidPhone),
		errors.Is(err, user.ErrEmptyPasswordHash):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("auth handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
