package handler

import (
	"context"
	"errors"
	"net/http"

	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/lessoncompletion"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// LessonCompletionService là tập con method của
// *lessoncompletionservice.Service mà handler cần (US4.10).
type LessonCompletionService interface {
	ListForStudent(ctx context.Context, courseID, userID uint, role user.Role) ([]lessoncompletion.LessonState, error)
	MarkComplete(ctx context.Context, lessonID, studentID uint) error
}

type LessonCompletionHandler struct {
	service LessonCompletionService
	log     *zap.Logger
}

func NewLessonCompletionHandler(service LessonCompletionService, log *zap.Logger) *LessonCompletionHandler {
	return &LessonCompletionHandler{service: service, log: log}
}

type lessonStateResponse struct {
	LessonID  uint `json:"lesson_id"`
	Completed bool `json:"completed"`
	Locked    bool `json:"locked"`
}

// ListForCourse xử lý GET /api/courses/{id}/lesson-progress (US4.10): trạng
// thái hoàn thành + khóa của mọi bài học trong khóa học, theo góc nhìn
// người gọi hiện tại (học viên thấy đúng tiến độ của mình; giảng viên/admin
// luôn thấy mở khóa hết).
func (h *LessonCompletionHandler) ListForCourse(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	courseID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "course id không hợp lệ")
		return
	}
	states, err := h.service.ListForStudent(r.Context(), courseID, claims.UserID, claims.Role)
	if err != nil {
		h.handleError(w, err)
		return
	}
	out := make([]lessonStateResponse, 0, len(states))
	for _, st := range states {
		out = append(out, lessonStateResponse{LessonID: st.LessonID, Completed: st.Completed, Locked: st.Locked})
	}
	writeJSON(w, http.StatusOK, out)
}

// MarkComplete xử lý POST /api/lessons/{id}/complete (US4.10, học viên).
func (h *LessonCompletionHandler) MarkComplete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	lessonID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "lesson id không hợp lệ")
		return
	}
	if err := h.service.MarkComplete(r.Context(), lessonID, claims.UserID); err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

func (h *LessonCompletionHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, curriculum.ErrLessonNotFound), errors.Is(err, curriculum.ErrChapterNotFound):
		writeError(w, http.StatusNotFound, "không tìm thấy hoặc bạn không có quyền truy cập")
	case errors.Is(err, lessoncompletion.ErrLessonLocked):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		h.log.Error("lesson completion handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
