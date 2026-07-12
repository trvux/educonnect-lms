package handler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/material"

	"go.uber.org/zap"
)

const maxUploadSize = 20 << 20 // 20MB, đủ cho slide/PDF bài giảng

// MaterialService là tập con method của *materialservice.Service mà
// handler cần (US4.1, US4.2).
type MaterialService interface {
	Upload(ctx context.Context, lessonID uint, fileName string, content io.Reader) (*material.Material, error)
	ListByLesson(ctx context.Context, lessonID uint) ([]*material.Material, error)
}

type MaterialHandler struct {
	service MaterialService
	log     *zap.Logger
}

func NewMaterialHandler(service MaterialService, log *zap.Logger) *MaterialHandler {
	return &MaterialHandler{service: service, log: log}
}

type materialResponse struct {
	ID       uint   `json:"id"`
	LessonID uint   `json:"lesson_id"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
}

func toMaterialResponse(m *material.Material) materialResponse {
	return materialResponse{ID: m.ID(), LessonID: m.LessonID(), FileName: m.FileName(), FilePath: m.FilePath()}
}

// Upload xử lý POST /api/lessons/{id}/materials, multipart/form-data với
// field "file" (US4.1, teacher-only).
func (h *MaterialHandler) Upload(w http.ResponseWriter, r *http.Request) {
	lessonID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "lesson id không hợp lệ")
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "file tải lên quá lớn hoặc không đúng định dạng multipart")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "thiếu file trong field \"file\"")
		return
	}
	defer file.Close()

	m, err := h.service.Upload(r.Context(), lessonID, header.Filename, file)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toMaterialResponse(m))
}

// List xử lý GET /api/lessons/{id}/materials (US4.2, học viên xem/tải tài liệu).
func (h *MaterialHandler) List(w http.ResponseWriter, r *http.Request) {
	lessonID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "lesson id không hợp lệ")
		return
	}
	materials, err := h.service.ListByLesson(r.Context(), lessonID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	out := make([]materialResponse, 0, len(materials))
	for _, m := range materials {
		out = append(out, toMaterialResponse(m))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *MaterialHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, curriculum.ErrLessonNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, material.ErrEmptyFileName), errors.Is(err, material.ErrInvalidLessonID):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("material handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
