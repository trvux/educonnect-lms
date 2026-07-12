package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"educonnect-lms/backend/internal/domain/curriculum"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// CurriculumService là tập con method của *curriculumservice.Service mà
// handler cần (US2.2).
type CurriculumService interface {
	CreateChapter(ctx context.Context, courseID uint, title string) (*curriculum.Chapter, error)
	ListChapters(ctx context.Context, courseID uint) ([]*curriculum.Chapter, error)
	CreateLesson(ctx context.Context, chapterID uint, title string) (*curriculum.Lesson, error)
	ListLessons(ctx context.Context, chapterID uint) ([]*curriculum.Lesson, error)
}

type CurriculumHandler struct {
	service CurriculumService
	log     *zap.Logger
}

func NewCurriculumHandler(service CurriculumService, log *zap.Logger) *CurriculumHandler {
	return &CurriculumHandler{service: service, log: log}
}

type titleRequest struct {
	Title string `json:"title"`
}

type chapterResponse struct {
	ID       uint   `json:"id"`
	CourseID uint   `json:"course_id"`
	Title    string `json:"title"`
	Position int    `json:"position"`
}

func toChapterResponse(c *curriculum.Chapter) chapterResponse {
	return chapterResponse{ID: c.ID(), CourseID: c.CourseID(), Title: c.Title(), Position: c.Position()}
}

type lessonResponse struct {
	ID        uint   `json:"id"`
	ChapterID uint   `json:"chapter_id"`
	Title     string `json:"title"`
	Position  int    `json:"position"`
}

func toLessonResponse(l *curriculum.Lesson) lessonResponse {
	return lessonResponse{ID: l.ID(), ChapterID: l.ChapterID(), Title: l.Title(), Position: l.Position()}
}

// CreateChapter xử lý POST /api/courses/{courseId}/chapters (US2.2, teacher-only).
func (h *CurriculumHandler) CreateChapter(w http.ResponseWriter, r *http.Request) {
	courseID, err := parseUintParam(r, "courseId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "course id không hợp lệ")
		return
	}
	var req titleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON không hợp lệ")
		return
	}
	ch, err := h.service.CreateChapter(r.Context(), courseID, req.Title)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toChapterResponse(ch))
}

// ListChapters xử lý GET /api/courses/{courseId}/chapters (public).
func (h *CurriculumHandler) ListChapters(w http.ResponseWriter, r *http.Request) {
	courseID, err := parseUintParam(r, "courseId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "course id không hợp lệ")
		return
	}
	chapters, err := h.service.ListChapters(r.Context(), courseID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	out := make([]chapterResponse, 0, len(chapters))
	for _, c := range chapters {
		out = append(out, toChapterResponse(c))
	}
	writeJSON(w, http.StatusOK, out)
}

// CreateLesson xử lý POST /api/chapters/{chapterId}/lessons (US2.2, teacher-only).
func (h *CurriculumHandler) CreateLesson(w http.ResponseWriter, r *http.Request) {
	chapterID, err := parseUintParam(r, "chapterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "chapter id không hợp lệ")
		return
	}
	var req titleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON không hợp lệ")
		return
	}
	l, err := h.service.CreateLesson(r.Context(), chapterID, req.Title)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toLessonResponse(l))
}

// ListLessons xử lý GET /api/chapters/{chapterId}/lessons (public).
func (h *CurriculumHandler) ListLessons(w http.ResponseWriter, r *http.Request) {
	chapterID, err := parseUintParam(r, "chapterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "chapter id không hợp lệ")
		return
	}
	lessons, err := h.service.ListLessons(r.Context(), chapterID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	out := make([]lessonResponse, 0, len(lessons))
	for _, l := range lessons {
		out = append(out, toLessonResponse(l))
	}
	writeJSON(w, http.StatusOK, out)
}

func parseUintParam(r *http.Request, name string) (uint, error) {
	return parseIDString(chi.URLParam(r, name))
}

func (h *CurriculumHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, curriculum.ErrChapterNotFound), errors.Is(err, curriculum.ErrLessonNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, curriculum.ErrEmptyChapterTitle), errors.Is(err, curriculum.ErrEmptyLessonTitle),
		errors.Is(err, curriculum.ErrInvalidCourseID), errors.Is(err, curriculum.ErrInvalidChapterID):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("curriculum handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
