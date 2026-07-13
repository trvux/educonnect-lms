package handler

import (
	"context"
	"net/http"

	"educonnect-lms/backend/internal/domain/report"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/handler/middleware"

	"go.uber.org/zap"
)

// ReportService là tập con method của *reportservice.Service mà handler
// cần (US7.2).
type ReportService interface {
	ForTeacher(ctx context.Context, teacherID uint) ([]report.CourseStats, error)
	All(ctx context.Context) ([]report.CourseStats, error)
}

type ReportHandler struct {
	service ReportService
	log     *zap.Logger
}

func NewReportHandler(service ReportService, log *zap.Logger) *ReportHandler {
	return &ReportHandler{service: service, log: log}
}

type courseStatsResponse struct {
	CourseID          uint    `json:"course_id"`
	CourseTitle       string  `json:"course_title"`
	EnrolledStudents  int     `json:"enrolled_students"`
	TotalAssignments  int     `json:"total_assignments"`
	AverageCompletion float64 `json:"average_completion"`
}

// Courses xử lý GET /api/reports/courses (US7.2). Quản trị viên xem thống
// kê toàn hệ thống; giảng viên chỉ xem thống kê các khóa học mình sở hữu.
func (h *ReportHandler) Courses(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "thiếu thông tin xác thực")
		return
	}

	var (
		stats []report.CourseStats
		err   error
	)
	if claims.Role == user.RoleAdmin {
		stats, err = h.service.All(r.Context())
	} else {
		stats, err = h.service.ForTeacher(r.Context(), claims.UserID)
	}
	if err != nil {
		h.log.Error("report handler: lỗi không xác định", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "lỗi hệ thống, vui lòng thử lại sau")
		return
	}

	out := make([]courseStatsResponse, 0, len(stats))
	for _, s := range stats {
		out = append(out, courseStatsResponse{
			CourseID:          s.CourseID,
			CourseTitle:       s.CourseTitle,
			EnrolledStudents:  s.EnrolledStudents,
			TotalAssignments:  s.TotalAssignments,
			AverageCompletion: s.AverageCompletion,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
