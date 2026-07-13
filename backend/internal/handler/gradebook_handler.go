package handler

import (
	"context"
	"errors"
	"net/http"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/gradebook"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// CourseGetter là tập con method của *courseservice.Service mà handler cần
// để xác nhận giảng viên xem bảng điểm chính là chủ sở hữu khóa học.
type CourseGetter interface {
	Get(ctx context.Context, id uint) (*course.Course, error)
}

// GradebookService là tập con method của *gradebookservice.Service.
type GradebookService interface {
	ForCourse(ctx context.Context, courseID uint) ([]gradebook.Entry, error)
}

type GradebookHandler struct {
	service GradebookService
	courses CourseGetter
	log     *zap.Logger
}

func NewGradebookHandler(service GradebookService, courses CourseGetter, log *zap.Logger) *GradebookHandler {
	return &GradebookHandler{service: service, courses: courses, log: log}
}

type gradebookEntryResponse struct {
	StudentID       uint     `json:"student_id"`
	StudentName     string   `json:"student_name"`
	AssignmentID    uint     `json:"assignment_id"`
	AssignmentTitle string   `json:"assignment_title"`
	Score           *float64 `json:"score,omitempty"`
}

// ForCourse xử lý GET /api/courses/{id}/gradebook (US5.3, chỉ giảng viên sở
// hữu khóa học hoặc quản trị viên).
func (h *GradebookHandler) ForCourse(w http.ResponseWriter, r *http.Request) {
	courseID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "course id không hợp lệ")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}

	c, err := h.courses.Get(r.Context(), courseID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	if claims.Role != user.RoleAdmin && c.TeacherID() != claims.UserID {
		writeError(w, http.StatusNotFound, course.ErrNotFound.Error()) // không tiết lộ sự tồn tại của khóa học người khác
		return
	}

	entries, err := h.service.ForCourse(r.Context(), courseID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	out := make([]gradebookEntryResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, gradebookEntryResponse{
			StudentID:       e.StudentID,
			StudentName:     e.StudentName,
			AssignmentID:    e.AssignmentID,
			AssignmentTitle: e.AssignmentTitle,
			Score:           e.Score,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *GradebookHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, course.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		h.log.Error("gradebook handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
