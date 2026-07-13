package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/submission"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// SubmissionService là tập con method của *submissionservice.Service mà
// handler cần (US5.2).
type SubmissionService interface {
	Submit(ctx context.Context, assignmentID, studentID uint, content string, answers []int) (*submission.Submission, error)
}

type SubmissionHandler struct {
	service SubmissionService
	log     *zap.Logger
}

func NewSubmissionHandler(service SubmissionService, log *zap.Logger) *SubmissionHandler {
	return &SubmissionHandler{service: service, log: log}
}

type submitRequest struct {
	Content string `json:"content"`
	Answers []int  `json:"answers"`
}

type submissionResponse struct {
	ID           uint   `json:"id"`
	AssignmentID uint   `json:"assignment_id"`
	StudentID    uint   `json:"student_id"`
	Content      string `json:"content"`
	Answers      []int  `json:"answers"`
}

func toSubmissionResponse(s *submission.Submission) submissionResponse {
	return submissionResponse{
		ID:           s.ID(),
		AssignmentID: s.AssignmentID(),
		StudentID:    s.StudentID(),
		Content:      s.Content(),
		Answers:      s.Answers(),
	}
}

// Submit xử lý POST /api/assignments/{id}/submit (US5.2, chỉ học viên đã
// đăng nhập — student id lấy từ JWT, không tin dữ liệu client gửi lên).
func (h *SubmissionHandler) Submit(w http.ResponseWriter, r *http.Request) {
	assignmentID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "assignment id không hợp lệ")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}

	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}

	s, err := h.service.Submit(r.Context(), assignmentID, claims.UserID, req.Content, req.Answers)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toSubmissionResponse(s))
}

func (h *SubmissionHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, assignment.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, submission.ErrAlreadySubmitted):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, submission.ErrPastDue),
		errors.Is(err, submission.ErrAnswerCountMismatch),
		errors.Is(err, submission.ErrEmptySubmission):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("submission handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
