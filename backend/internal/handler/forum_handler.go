package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/forum"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// ForumService là tập con method của *forumservice.Service mà handler cần
// (US6.1).
type ForumService interface {
	Post(ctx context.Context, courseID, authorID uint, parentID *uint, content string) (*forum.Post, error)
	ListByCourse(ctx context.Context, courseID uint) ([]*forum.Post, error)
}

type ForumHandler struct {
	service ForumService
	log     *zap.Logger
}

func NewForumHandler(service ForumService, log *zap.Logger) *ForumHandler {
	return &ForumHandler{service: service, log: log}
}

type createPostRequest struct {
	ParentID *uint  `json:"parent_id"`
	Content  string `json:"content"`
}

type postResponse struct {
	ID         uint      `json:"id"`
	CourseID   uint      `json:"course_id"`
	AuthorID   uint      `json:"author_id"`
	AuthorName string    `json:"author_name,omitempty"`
	ParentID   *uint     `json:"parent_id,omitempty"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

func toPostResponse(p *forum.Post) postResponse {
	return postResponse{
		ID:         p.ID(),
		CourseID:   p.CourseID(),
		AuthorID:   p.AuthorID(),
		AuthorName: p.AuthorName(),
		ParentID:   p.ParentID(),
		Content:    p.Content(),
		CreatedAt:  p.CreatedAt(),
	}
}

// Create xử lý POST /api/courses/{id}/forum-posts (US6.1, mọi user đã
// đăng nhập — học viên hoặc giảng viên).
func (h *ForumHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dữ liệu gửi lên không đúng định dạng JSON")
		return
	}

	p, err := h.service.Post(r.Context(), courseID, claims.UserID, req.ParentID, req.Content)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toPostResponse(p))
}

// List xử lý GET /api/courses/{id}/forum-posts (US6.1, public).
func (h *ForumHandler) List(w http.ResponseWriter, r *http.Request) {
	courseID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "course id không hợp lệ")
		return
	}
	posts, err := h.service.ListByCourse(r.Context(), courseID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	out := make([]postResponse, 0, len(posts))
	for _, p := range posts {
		out = append(out, toPostResponse(p))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ForumHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, course.ErrNotFound), errors.Is(err, forum.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, forum.ErrEmptyContent),
		errors.Is(err, forum.ErrInvalidCourseID),
		errors.Is(err, forum.ErrInvalidAuthorID),
		errors.Is(err, forum.ErrParentCourseMismatch):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("forum handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
