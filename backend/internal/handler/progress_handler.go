package handler

import (
	"context"
	"net/http"

	"educonnect-lms/backend/internal/domain/progress"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// ProgressService là tập con method của *progressservice.Service mà
// handler cần (US7.1).
type ProgressService interface {
	ForStudent(ctx context.Context, studentID uint) ([]progress.CourseProgress, error)
}

type ProgressHandler struct {
	service ProgressService
	log     *zap.Logger
}

func NewProgressHandler(service ProgressService, log *zap.Logger) *ProgressHandler {
	return &ProgressHandler{service: service, log: log}
}

type courseProgressResponse struct {
	CourseID         uint    `json:"course_id"`
	CourseTitle      string  `json:"course_title"`
	TotalAssignments int     `json:"total_assignments"`
	Submitted        int     `json:"submitted"`
	PercentComplete  float64 `json:"percent_complete"`
}

// Me xử lý GET /api/me/progress (US7.1, dashboard tiến độ của chính học
// viên đang đăng nhập).
func (h *ProgressHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}

	list, err := h.service.ForStudent(r.Context(), claims.UserID)
	if err != nil {
		h.log.Error("progress handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
		return
	}

	out := make([]courseProgressResponse, 0, len(list))
	for _, p := range list {
		out = append(out, courseProgressResponse{
			CourseID:         p.CourseID,
			CourseTitle:      p.CourseTitle,
			TotalAssignments: p.TotalAssignments,
			Submitted:        p.Submitted,
			PercentComplete:  p.PercentComplete,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
