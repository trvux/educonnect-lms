package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"educonnect-lms/backend/internal/domain/roleupgrade"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// RoleUpgradeService là tập con method của *roleupgradeservice.Service mà
// handler cần (US1.7).
type RoleUpgradeService interface {
	Create(ctx context.Context, userID uint, reason string) (*roleupgrade.Request, error)
	ListPending(ctx context.Context) ([]*roleupgrade.Request, error)
	Approve(ctx context.Context, requestID, adminID uint) (*roleupgrade.Request, error)
	Reject(ctx context.Context, requestID, adminID uint) (*roleupgrade.Request, error)
}

type RoleUpgradeHandler struct {
	service RoleUpgradeService
	log     *zap.Logger
}

func NewRoleUpgradeHandler(service RoleUpgradeService, log *zap.Logger) *RoleUpgradeHandler {
	return &RoleUpgradeHandler{service: service, log: log}
}

type roleUpgradeRequestResponse struct {
	ID         uint       `json:"id"`
	UserID     uint       `json:"user_id"`
	Reason     string     `json:"reason"`
	Status     string     `json:"status"`
	ReviewedBy *uint      `json:"reviewed_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
}

func toRoleUpgradeResponse(r *roleupgrade.Request) roleUpgradeRequestResponse {
	return roleUpgradeRequestResponse{
		ID:         r.ID(),
		UserID:     r.UserID(),
		Reason:     r.Reason(),
		Status:     string(r.Status()),
		ReviewedBy: r.ReviewedBy(),
		CreatedAt:  r.CreatedAt(),
		ReviewedAt: r.ReviewedAt(),
	}
}

type createRoleUpgradeRequest struct {
	Reason string `json:"reason"`
}

// Create xử lý POST /api/me/role-upgrade-request (US1.7, chỉ Student).
func (h *RoleUpgradeHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	var req createRoleUpgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}
	created, err := h.service.Create(r.Context(), claims.UserID, req.Reason)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toRoleUpgradeResponse(created))
}

// ListPending xử lý GET /api/admin/role-upgrade-requests (US1.7, admin).
func (h *RoleUpgradeHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListPending(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}
	out := make([]roleUpgradeRequestResponse, 0, len(list))
	for _, req := range list {
		out = append(out, toRoleUpgradeResponse(req))
	}
	writeJSON(w, http.StatusOK, out)
}

// Approve xử lý POST /api/admin/role-upgrade-requests/{id}/approve.
func (h *RoleUpgradeHandler) Approve(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	id, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "request id không hợp lệ")
		return
	}
	updated, err := h.service.Approve(r.Context(), id, claims.UserID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toRoleUpgradeResponse(updated))
}

// Reject xử lý POST /api/admin/role-upgrade-requests/{id}/reject.
func (h *RoleUpgradeHandler) Reject(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	id, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "request id không hợp lệ")
		return
	}
	updated, err := h.service.Reject(r.Context(), id, claims.UserID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toRoleUpgradeResponse(updated))
}

func (h *RoleUpgradeHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, roleupgrade.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, roleupgrade.ErrAlreadyPending):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, roleupgrade.ErrEmptyReason), errors.Is(err, roleupgrade.ErrNotPending):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("role upgrade handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
