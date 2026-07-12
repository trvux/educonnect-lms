package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/handler/middleware"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// CourseService is the subset of *course.Service the handler depends on.
type CourseService interface {
	Create(ctx context.Context, in courseCreateInput) (*course.Course, error)
	Search(ctx context.Context, keyword string) ([]*course.Course, error)
	SubmitForReview(ctx context.Context, courseID, teacherID uint) (*course.Course, error)
	Approve(ctx context.Context, courseID uint) (*course.Course, error)
}

// courseCreateInput mirrors service/course.CreateInput; declared locally so
// this package does not need to import the concrete service package just
// for the interface signature (kept satisfiable by a light adapter in router.go).
type courseCreateInput = struct {
	Title       string
	Description string
	TeacherID   uint
}

type CourseHandler struct {
	service CourseService
	log     *zap.Logger
}

func NewCourseHandler(service CourseService, log *zap.Logger) *CourseHandler {
	return &CourseHandler{service: service, log: log}
}

type createCourseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type courseResponse struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	TeacherID   uint   `json:"teacher_id"`
}

func toCourseResponse(c *course.Course) courseResponse {
	return courseResponse{
		ID: c.ID(), Title: c.Title(), Description: c.Description(),
		Status: string(c.Status()), TeacherID: c.TeacherID(),
	}
}

// Create handles POST /api/courses (US2.1, teacher-only — see router.go).
func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing auth context")
		return
	}

	var req createCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	c, err := h.service.Create(r.Context(), courseCreateInput{
		Title: req.Title, Description: req.Description, TeacherID: claims.UserID,
	})
	if err != nil {
		h.handleCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toCourseResponse(c))
}

// Search handles GET /api/courses?search=... (US3.1, public).
func (h *CourseHandler) Search(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("search")
	results, err := h.service.Search(r.Context(), keyword)
	if err != nil {
		h.log.Error("course handler: search failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	out := make([]courseResponse, 0, len(results))
	for _, c := range results {
		out = append(out, toCourseResponse(c))
	}
	writeJSON(w, http.StatusOK, out)
}

// SubmitForReview handles POST /api/courses/{id}/submit (teacher-only).
func (h *CourseHandler) SubmitForReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing auth context")
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}
	c, err := h.service.SubmitForReview(r.Context(), id, claims.UserID)
	if err != nil {
		h.handleCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCourseResponse(c))
}

// Approve handles POST /api/admin/courses/{id}/approve (US2.3, admin-only).
func (h *CourseHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}
	c, err := h.service.Approve(r.Context(), id)
	if err != nil {
		h.handleCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCourseResponse(c))
}

func parseIDParam(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (h *CourseHandler) handleCourseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, course.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, course.ErrNotPending):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, course.ErrEmptyTitle), errors.Is(err, course.ErrInvalidTeacherID):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("course handler: unexpected error", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
