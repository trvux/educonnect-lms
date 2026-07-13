package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/notification"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// NotificationService là tập con method của *notificationservice.Service
// mà handler cần (US6.2).
type NotificationService interface {
	SendToCourse(ctx context.Context, courseID uint, title, message string) ([]*notification.Notification, error)
	ListMine(ctx context.Context, recipientID uint) ([]*notification.Notification, error)
	UnreadCount(ctx context.Context, recipientID uint) (int, error)
	MarkRead(ctx context.Context, id, recipientID uint) error
}

type NotificationHandler struct {
	service NotificationService
	log     *zap.Logger
}

func NewNotificationHandler(service NotificationService, log *zap.Logger) *NotificationHandler {
	return &NotificationHandler{service: service, log: log}
}

type sendNotificationRequest struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type notificationResponse struct {
	ID        uint      `json:"id"`
	CourseID  uint      `json:"course_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

func toNotificationResponse(n *notification.Notification) notificationResponse {
	return notificationResponse{
		ID:        n.ID(),
		CourseID:  n.CourseID(),
		Title:     n.Title(),
		Message:   n.Message(),
		Read:      n.Read(),
		CreatedAt: n.CreatedAt(),
	}
}

// SendToCourse xử lý POST /api/courses/{id}/notifications (US6.2, chỉ
// giảng viên/quản trị viên).
func (h *NotificationHandler) SendToCourse(w http.ResponseWriter, r *http.Request) {
	courseID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "course id không hợp lệ")
		return
	}

	var req sendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}

	sent, err := h.service.SendToCourse(r.Context(), courseID, req.Title, req.Message)
	if err != nil {
		h.handleError(w, err)
		return
	}
	out := make([]notificationResponse, 0, len(sent))
	for _, n := range sent {
		out = append(out, toNotificationResponse(n))
	}
	writeJSON(w, http.StatusCreated, out)
}

// ListMine xử lý GET /api/notifications (US6.2, thông báo của người dùng
// hiện tại — phục vụ chuông thông báo trên FE).
func (h *NotificationHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	list, err := h.service.ListMine(r.Context(), claims.UserID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	out := make([]notificationResponse, 0, len(list))
	for _, n := range list {
		out = append(out, toNotificationResponse(n))
	}
	writeJSON(w, http.StatusOK, out)
}

// UnreadCount xử lý GET /api/notifications/unread-count — badge số chưa
// đọc trên chuông thông báo.
func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	count, err := h.service.UnreadCount(r.Context(), claims.UserID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

// MarkRead xử lý POST /api/notifications/{id}/read.
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "notification id không hợp lệ")
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	if err := h.service.MarkRead(r.Context(), id, claims.UserID); err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

func (h *NotificationHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, course.ErrNotFound), errors.Is(err, notification.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, notification.ErrEmptyTitle),
		errors.Is(err, notification.ErrInvalidCourseID),
		errors.Is(err, notification.ErrInvalidRecipientID):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("notification handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
