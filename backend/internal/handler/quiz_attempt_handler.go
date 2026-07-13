package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/quizattempt"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// QuizAttemptService là tập con method của *quizattemptservice.Service mà
// handler cần (US5.4).
type QuizAttemptService interface {
	Start(ctx context.Context, assignmentID, studentID uint) (*quizattempt.QuizAttempt, error)
}

type QuizAttemptHandler struct {
	service QuizAttemptService
	log     *zap.Logger
}

func NewQuizAttemptHandler(service QuizAttemptService, log *zap.Logger) *QuizAttemptHandler {
	return &QuizAttemptHandler{service: service, log: log}
}

type quizAttemptResponse struct {
	StartedAt time.Time `json:"started_at"`
}

// Start xử lý POST /api/assignments/{id}/start-attempt (US5.4, học viên):
// ghi nhận thời điểm bắt đầu làm bài trắc nghiệm — idempotent, gọi lại
// (refresh trang) trả về đúng started_at gốc, không reset đồng hồ đếm
// ngược. Frontend tự tính deadline = started_at + time_limit_minutes (lấy
// từ GET /assignments/{id}) — handler này không cần biết time_limit_minutes.
func (h *QuizAttemptHandler) Start(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	assignmentID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "assignment id không hợp lệ")
		return
	}
	attempt, err := h.service.Start(r.Context(), assignmentID, claims.UserID)
	if err != nil {
		if errors.Is(err, assignment.ErrNotFound) {
			writeError(w, http.StatusNotFound, "không tìm thấy bài tập")
			return
		}
		h.log.Error("quiz attempt handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
		return
	}
	writeJSON(w, http.StatusOK, quizAttemptResponse{StartedAt: attempt.StartedAt()})
}
