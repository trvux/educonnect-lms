package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/handler/middleware"
	courseservice "educonnect-lms/backend/internal/service/course"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// CourseService là tập con method của *courseservice.Service mà handler
// cần — để dạng interface để test có thể inject fake.
type CourseService interface {
	Create(ctx context.Context, in courseservice.CreateInput) (*course.Course, error)
	Get(ctx context.Context, id uint) (*course.Course, error)
	Search(ctx context.Context, keyword string) ([]*course.Course, error)
	ListPending(ctx context.Context) ([]*course.Course, error)
	SubmitForReview(ctx context.Context, courseID, teacherID uint) (*course.Course, error)
	Approve(ctx context.Context, courseID uint) (*course.Course, error)
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

// Create xử lý POST /api/courses (US2.1, chỉ teacher — xem router.go).
func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}

	var req createCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON không hợp lệ")
		return
	}

	c, err := h.service.Create(r.Context(), courseservice.CreateInput{
		Title: req.Title, Description: req.Description, TeacherID: claims.UserID,
	})
	if err != nil {
		h.handleCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toCourseResponse(c))
}

// Search xử lý GET /api/courses?search=... (US3.1, public).
func (h *CourseHandler) Search(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("search")
	results, err := h.service.Search(r.Context(), keyword)
	if err != nil {
		h.log.Error("course handler: tìm kiếm thất bại", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
		return
	}
	out := make([]courseResponse, 0, len(results))
	for _, c := range results {
		out = append(out, toCourseResponse(c))
	}
	writeJSON(w, http.StatusOK, out)
}

// Get xử lý GET /api/courses/{id} (public — xem chi tiết 1 khóa học).
func (h *CourseHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "course id không hợp lệ")
		return
	}
	c, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.handleCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCourseResponse(c))
}

// ListPending xử lý GET /api/admin/courses/pending (US2.3, chỉ admin —
// hàng chờ duyệt).
func (h *CourseHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	results, err := h.service.ListPending(r.Context())
	if err != nil {
		h.log.Error("course handler: lấy hàng chờ duyệt thất bại", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
		return
	}
	out := make([]courseResponse, 0, len(results))
	for _, c := range results {
		out = append(out, toCourseResponse(c))
	}
	writeJSON(w, http.StatusOK, out)
}

// SubmitForReview xử lý POST /api/courses/{id}/submit (chỉ teacher).
func (h *CourseHandler) SubmitForReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "course id không hợp lệ")
		return
	}
	c, err := h.service.SubmitForReview(r.Context(), id, claims.UserID)
	if err != nil {
		h.handleCourseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCourseResponse(c))
}

// Approve xử lý POST /api/admin/courses/{id}/approve (US2.3, chỉ admin).
func (h *CourseHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "course id không hợp lệ")
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
	return parseIDString(chi.URLParam(r, "id"))
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
		h.log.Error("course handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
