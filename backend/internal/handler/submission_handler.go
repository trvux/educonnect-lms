package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/submission"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// SubmissionService là tập con method của *submissionservice.Service mà
// handler cần (US5.2, US5.3).
type SubmissionService interface {
	Submit(ctx context.Context, assignmentID, studentID uint, content string, answers []int) (*submission.Submission, error)
	Grade(ctx context.Context, submissionID, graderID uint, isAdmin bool, score float64, feedback string) (*submission.Submission, error)
	ListByAssignment(ctx context.Context, assignmentID uint) ([]*submission.Submission, error)
	GetMine(ctx context.Context, assignmentID, studentID uint) (*submission.Submission, error)
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

type gradeRequest struct {
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
}

type submissionResponse struct {
	ID           uint     `json:"id"`
	AssignmentID uint     `json:"assignment_id"`
	StudentID    uint     `json:"student_id"`
	Content      string   `json:"content"`
	Answers      []int    `json:"answers"`
	Score        *float64 `json:"score,omitempty"`
	Feedback     string   `json:"feedback,omitempty"`
	Graded       bool     `json:"graded"`
}

func toSubmissionResponse(s *submission.Submission) submissionResponse {
	return submissionResponse{
		ID:           s.ID(),
		AssignmentID: s.AssignmentID(),
		StudentID:    s.StudentID(),
		Content:      s.Content(),
		Answers:      s.Answers(),
		Score:        s.Score(),
		Feedback:     s.Feedback(),
		Graded:       s.IsGraded(),
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

// Grade xử lý POST /api/submissions/{id}/grade (US5.3, chỉ giảng viên sở
// hữu khóa học hoặc quản trị viên).
func (h *SubmissionHandler) Grade(w http.ResponseWriter, r *http.Request) {
	submissionID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "submission id không hợp lệ")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}

	var req gradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}

	s, err := h.service.Grade(r.Context(), submissionID, claims.UserID, claims.Role == user.RoleAdmin, req.Score, req.Feedback)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSubmissionResponse(s))
}

// ListByAssignment xử lý GET /api/assignments/{id}/submissions (US5.3,
// giảng viên/quản trị viên xem danh sách bài nộp để chấm điểm).
func (h *SubmissionHandler) ListByAssignment(w http.ResponseWriter, r *http.Request) {
	assignmentID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "assignment id không hợp lệ")
		return
	}
	submissions, err := h.service.ListByAssignment(r.Context(), assignmentID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	out := make([]submissionResponse, 0, len(submissions))
	for _, s := range submissions {
		out = append(out, toSubmissionResponse(s))
	}
	writeJSON(w, http.StatusOK, out)
}

// GetMine xử lý GET /api/assignments/{id}/my-submission (chỉ học viên đã
// đăng nhập — trả về ErrNotFound nếu chưa nộp). Cho phép FE biết ngay
// trạng thái đã nộp/điểm khi vào trang làm bài, không cần đợi bấm Nộp bài.
func (h *SubmissionHandler) GetMine(w http.ResponseWriter, r *http.Request) {
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

	s, err := h.service.GetMine(r.Context(), assignmentID, claims.UserID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSubmissionResponse(s))
}

func (h *SubmissionHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, assignment.ErrNotFound), errors.Is(err, submission.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, submission.ErrAlreadySubmitted):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, submission.ErrPastDue),
		errors.Is(err, submission.ErrAnswerCountMismatch),
		errors.Is(err, submission.ErrEmptySubmission),
		errors.Is(err, submission.ErrInvalidScore):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("submission handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
