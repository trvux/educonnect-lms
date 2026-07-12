package handler

import (
	"context"
	"errors"
	"net/http"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/enrollment"
	"educonnect-lms/backend/internal/handler/middleware"
	enrollmentservice "educonnect-lms/backend/internal/service/enrollment"

	"go.uber.org/zap"
)

// EnrollmentService là tập con method của *enrollmentservice.Service mà
// handler cần (US3.2, US3.3).
type EnrollmentService interface {
	Enroll(ctx context.Context, studentID, courseID uint) (*enrollment.Enrollment, error)
	ListStudents(ctx context.Context, courseID, requestingTeacherID uint) ([]enrollmentservice.EnrolledStudent, error)
}

type EnrollmentHandler struct {
	service EnrollmentService
	log     *zap.Logger
}

func NewEnrollmentHandler(service EnrollmentService, log *zap.Logger) *EnrollmentHandler {
	return &EnrollmentHandler{service: service, log: log}
}

type enrollmentResponse struct {
	ID         uint   `json:"id"`
	CourseID   uint   `json:"course_id"`
	StudentID  uint   `json:"student_id"`
	EnrolledAt string `json:"enrolled_at"`
}

// Enroll xử lý POST /api/courses/{id}/enroll (US3.2, cần đăng nhập — mọi
// role đã xác thực đều có thể đăng ký, thường dùng bởi student).
func (h *EnrollmentHandler) Enroll(w http.ResponseWriter, r *http.Request) {
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

	e, err := h.service.Enroll(r.Context(), claims.UserID, courseID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, enrollmentResponse{
		ID: e.ID(), CourseID: e.CourseID(), StudentID: e.StudentID(),
		EnrolledAt: e.EnrolledAt().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListStudents xử lý GET /api/courses/{id}/students (US3.3, teacher-only —
// service tự kiểm tra chỉ giáo viên sở hữu khóa học mới xem được).
func (h *EnrollmentHandler) ListStudents(w http.ResponseWriter, r *http.Request) {
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

	students, err := h.service.ListStudents(r.Context(), courseID, claims.UserID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, students)
}

func (h *EnrollmentHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, course.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, enrollment.ErrAlreadyEnrolled):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, enrollmentservice.ErrCourseNotApproved):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, enrollment.ErrInvalidStudentID), errors.Is(err, enrollment.ErrInvalidCourseID):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.log.Error("enrollment handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
	}
}
