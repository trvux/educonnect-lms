package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// AssignmentService là tập con method của *assignmentservice.Service mà
// handler cần (US5.1).
type AssignmentService interface {
	Create(ctx context.Context, lessonID uint, title, description string, kind assignment.Type, questions []assignment.Question, dueAt *time.Time) (*assignment.Assignment, error)
	Get(ctx context.Context, id uint) (*assignment.Assignment, error)
	ListByLesson(ctx context.Context, lessonID uint) ([]*assignment.Assignment, error)
}

type AssignmentHandler struct {
	service AssignmentService
	log     *zap.Logger
}

func NewAssignmentHandler(service AssignmentService, log *zap.Logger) *AssignmentHandler {
	return &AssignmentHandler{service: service, log: log}
}

type questionRequest struct {
	Content      string   `json:"content"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
}

type createAssignmentRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Kind        string            `json:"kind"`
	Questions   []questionRequest `json:"questions"`
	DueAt       *time.Time        `json:"due_at"`
}

type questionResponse struct {
	Content string   `json:"content"`
	Options []string `json:"options"`
	// CorrectIndex chỉ hiển thị cho giảng viên/quản trị viên — ẩn với học
	// viên để tránh lộ đáp án trước khi nộp bài (US5.2, US5.3).
	CorrectIndex *int `json:"correct_index,omitempty"`
}

type assignmentResponse struct {
	ID          uint               `json:"id"`
	LessonID    uint               `json:"lesson_id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Kind        string             `json:"kind"`
	Questions   []questionResponse `json:"questions,omitempty"`
	DueAt       *time.Time         `json:"due_at,omitempty"`
}

// canSeeAnswers chỉ giảng viên/quản trị viên đã đăng nhập mới xem được đáp
// án đúng của câu hỏi trắc nghiệm.
func canSeeAnswers(ctx context.Context) bool {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return false
	}
	return claims.Role == user.RoleTeacher || claims.Role == user.RoleAdmin
}

func toAssignmentResponse(a *assignment.Assignment, showAnswers bool) assignmentResponse {
	questions := make([]questionResponse, 0, len(a.Questions()))
	for _, q := range a.Questions() {
		qr := questionResponse{Content: q.Content, Options: q.Options}
		if showAnswers {
			idx := q.CorrectIndex
			qr.CorrectIndex = &idx
		}
		questions = append(questions, qr)
	}
	return assignmentResponse{
		ID:          a.ID(),
		LessonID:    a.LessonID(),
		Title:       a.Title(),
		Description: a.Description(),
		Kind:        string(a.Kind()),
		Questions:   questions,
		DueAt:       a.DueAt(),
	}
}

// Create xử lý POST /api/lessons/{id}/assignments (US5.1, teacher/admin only).
func (h *AssignmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	lessonID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "lesson id không hợp lệ")
		return
	}

	var req createAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}

	questions := make([]assignment.Question, len(req.Questions))
	for i, q := range req.Questions {
		questions[i] = assignment.Question{Content: q.Content, Options: q.Options, CorrectIndex: q.CorrectIndex}
	}

	a, err := h.service.Create(r.Context(), lessonID, req.Title, req.Description, assignment.Type(req.Kind), questions, req.DueAt)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAssignmentResponse(a, true))
}

// List xử lý GET /api/lessons/{id}/assignments (US5.1, public — ẩn đáp án
// đúng nếu người gọi không phải giảng viên/quản trị viên).
func (h *AssignmentHandler) List(w http.ResponseWriter, r *http.Request) {
	lessonID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "lesson id không hợp lệ")
		return
	}
	assignments, err := h.service.ListByLesson(r.Context(), lessonID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	showAnswers := canSeeAnswers(r.Context())
	out := make([]assignmentResponse, 0, len(assignments))
	for _, a := range assignments {
		out = append(out, toAssignmentResponse(a, showAnswers))
	}
	writeJSON(w, http.StatusOK, out)
}

// Get xử lý GET /api/assignments/{id} (US5.1, public — ẩn đáp án đúng nếu
// người gọi không phải giảng viên/quản trị viên).
func (h *AssignmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "assignment id không hợp lệ")
		return
	}
	a, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAssignmentResponse(a, canSeeAnswers(r.Context())))
}

func (h *AssignmentHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, curriculum.ErrLessonNotFound), errors.Is(err, assignment.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, assignment.ErrEmptyTitle),
		errors.Is(err, assignment.ErrInvalidLessonID),
		errors.Is(err, assignment.ErrInvalidType),
		errors.Is(err, assignment.ErrQuizNeedsQuestions),
		errors.Is(err, assignment.ErrInvalidQuestion):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("assignment handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
