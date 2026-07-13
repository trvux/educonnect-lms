package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/material"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

const maxUploadSize = 20 << 20 // 20MB, đủ cho slide/PDF bài giảng

// MaterialService là tập con method của *materialservice.Service mà
// handler cần (US4.1, US4.2, US4.3).
type MaterialService interface {
	Upload(ctx context.Context, lessonID uint, fileName string, content io.Reader) (*material.Material, error)
	ListByLesson(ctx context.Context, lessonID, userID uint, role user.Role) ([]*material.Material, error)
	Get(ctx context.Context, materialID, userID uint, role user.Role) (*material.Material, error)
	Delete(ctx context.Context, materialID, userID uint, role user.Role) error
}

type MaterialHandler struct {
	service MaterialService
	log     *zap.Logger
	// uploadsDir là thư mục gốc lưu file vật lý trên đĩa (US4.3: cần để mở
	// file thật khi phục vụ download, thay vì serve tĩnh qua http.FileServer
	// không kiểm tra quyền như trước).
	uploadsDir string
}

func NewMaterialHandler(service MaterialService, log *zap.Logger, uploadsDir string) *MaterialHandler {
	return &MaterialHandler{service: service, log: log, uploadsDir: uploadsDir}
}

type materialResponse struct {
	ID       uint   `json:"id"`
	LessonID uint   `json:"lesson_id"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	FileType string `json:"file_type"`
}

func toMaterialResponse(m *material.Material) materialResponse {
	return materialResponse{
		ID:       m.ID(),
		LessonID: m.LessonID(),
		FileName: m.FileName(),
		FilePath: m.FilePath(),
		FileType: string(m.FileType()),
	}
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

// List xử lý GET /api/lessons/{id}/materials (US4.2). US4.3: chỉ trả danh
// sách nếu người gọi đã đăng nhập và có quyền truy cập lesson (đã đăng ký
// khóa học, sở hữu khóa học, hoặc admin) — route này không còn public.
func (h *MaterialHandler) List(w http.ResponseWriter, r *http.Request) {
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
	materials, err := h.service.ListByLesson(r.Context(), lessonID, claims.UserID, claims.Role)
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

// Download xử lý GET /api/materials/{id}/download (US4.3). Thay cho việc
// serve tĩnh qua http.FileServer trên "/uploads/*" (không kiểm tra quyền
// gì cả — lỗ hổng bảo mật thật đã phát hiện), endpoint này kiểm tra quyền
// truy cập trước, rồi dùng http.ServeContent (hỗ trợ Range, để video/PDF
// tua được) kèm Content-Disposition: attachment để trình duyệt tải file
// về máy thay vì chỉ preview (sửa luôn lỗi "bấm download không ra gì").
func (h *MaterialHandler) Download(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	materialID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "material id không hợp lệ")
		return
	}
	m, err := h.service.Get(r.Context(), materialID, claims.UserID, claims.Role)
	if err != nil {
		h.handleError(w, err)
		return
	}

	fullPath := filepath.Join(h.uploadsDir, m.FilePath())
	f, err := os.Open(fullPath)
	if err != nil {
		h.log.Error("material handler: mở file lỗi", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		h.log.Error("material handler: đọc thông tin file lỗi", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, m.FileName()))
	http.ServeContent(w, r, m.FileName(), info.ModTime(), f)
}

// Delete xử lý DELETE /api/materials/{id} (US4.8, chỉ GV sở hữu khóa học
// hoặc admin — học viên không bao giờ được xóa).
func (h *MaterialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}
	materialID, err := parseUintParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "material id không hợp lệ")
		return
	}
	if err := h.service.Delete(r.Context(), materialID, claims.UserID, claims.Role); err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

func (h *MaterialHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, curriculum.ErrLessonNotFound), errors.Is(err, material.ErrNotFound):
		writeError(w, http.StatusNotFound, "không tìm thấy hoặc bạn không có quyền truy cập")
	case errors.Is(err, material.ErrEmptyFileName), errors.Is(err, material.ErrInvalidLessonID), errors.Is(err, material.ErrUnsupportedFileType):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("material handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
