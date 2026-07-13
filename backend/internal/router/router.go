package router

import (
	"net/http"
	"time"

	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/handler"
	"educonnect-lms/backend/internal/handler/middleware"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Deps struct {
	AuthHandler          *handler.AuthHandler
	CourseHandler        *handler.CourseHandler
	CurriculumHandler    *handler.CurriculumHandler
	EnrollmentHandler    *handler.EnrollmentHandler
	MaterialHandler      *handler.MaterialHandler
	AssignmentHandler    *handler.AssignmentHandler
	SubmissionHandler    *handler.SubmissionHandler
	GradebookHandler     *handler.GradebookHandler
	ForumHandler         *handler.ForumHandler
	NotificationHandler  *handler.NotificationHandler
	ProgressHandler      *handler.ProgressHandler
	ReportHandler        *handler.ReportHandler
	PasswordResetHandler *handler.PasswordResetHandler
	RoleUpgradeHandler   *handler.RoleUpgradeHandler
	TokenVerifier        middleware.TokenVerifier
	// StreamTokenVerifier xác thực token ngắn hạn riêng cho US4.5 (query
	// param "token", khác JWT đăng nhập dài hạn ở TokenVerifier).
	StreamTokenVerifier middleware.TokenVerifier
	// UploadsDir là thư mục lưu file vật lý (US4.1), phục vụ tĩnh qua
	// /uploads/* để frontend tải xuống (US4.2).
	UploadsDir string
}

func New(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	// Cho phép frontend Next.js (chạy port khác) gọi API — cần thiết vì
	// frontend/backend là 2 origin riêng biệt (không cùng domain/port).
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// US4.3: chỉ avatar (ảnh đại diện, không nhạy cảm) được phục vụ tĩnh
	// công khai. Tài liệu bài giảng KHÔNG còn serve qua /uploads/* nữa — lỗ
	// hổng bảo mật trước đây (ai có link cũng tải được, không cần đăng
	// nhập/đăng ký khóa học) đã được vá bằng cách bắt buộc đi qua endpoint
	// có kiểm tra quyền GET /api/materials/{id}/download bên dưới.
	uploadsDir := deps.UploadsDir
	if uploadsDir == "" {
		uploadsDir = "uploads"
	}
	avatarServer := http.StripPrefix("/uploads/avatars/", http.FileServer(http.Dir(uploadsDir+"/avatars")))
	r.Handle("/uploads/avatars/*", avatarServer)

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", deps.AuthHandler.Register)                      // US1.1
			r.Post("/login", deps.AuthHandler.Login)                            // US1.2
			r.Post("/verify-email", deps.AuthHandler.VerifyEmail)               // US1.9, public
			r.Post("/resend-verification", deps.AuthHandler.ResendVerification) // US1.9, public
			r.Post("/forgot-username", deps.AuthHandler.ForgotUsername)         // US1.8, public
			r.Post("/forgot-password", deps.PasswordResetHandler.Forgot)        // US1.6, public
			r.Post("/reset-password", deps.PasswordResetHandler.Reset)          // US1.6, public
		})

		r.Get("/courses", deps.CourseHandler.Search)                               // US3.1, public
		r.Get("/courses/{id}", deps.CourseHandler.Get)                             // xem chi tiết khóa học, public
		r.Get("/courses/{courseId}/chapters", deps.CurriculumHandler.ListChapters) // US2.2, public
		r.Get("/chapters/{chapterId}/lessons", deps.CurriculumHandler.ListLessons) // US2.2, public
		r.Get("/courses/{id}/forum-posts", deps.ForumHandler.List)                 // US6.1, public

		// OptionalAuth: route public nhưng cần biết vai trò người gọi để ẩn
		// đáp án đúng của trắc nghiệm với học viên/khách (US5.1).
		r.Group(func(r chi.Router) {
			r.Use(middleware.OptionalAuth(deps.TokenVerifier))
			r.Get("/lessons/{id}/assignments", deps.AssignmentHandler.List)
			r.Get("/assignments/{id}", deps.AssignmentHandler.Get)
		})

		// US4.5: nhóm route riêng cho <video src="...">, xác thực qua query
		// param "token" (RequireStreamAuth) thay vì header Authorization mà
		// thẻ video không gửi được.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireStreamAuth(deps.StreamTokenVerifier))
			r.Get("/materials/{id}/stream", deps.MaterialHandler.Stream)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(deps.TokenVerifier))

			r.Post("/courses/{id}/enroll", deps.EnrollmentHandler.Enroll) // US3.2, mọi user đã đăng nhập
			r.Post("/courses/{id}/forum-posts", deps.ForumHandler.Create) // US6.1, mọi user đã đăng nhập

			r.Get("/me", deps.AuthHandler.Me)                                // US1.4
			r.Patch("/me", deps.AuthHandler.UpdateMe)                        // US1.4
			r.Post("/me/avatar", deps.AuthHandler.UploadAvatar)              // US1.4
			r.Post("/auth/change-password", deps.AuthHandler.ChangePassword) // US1.5

			r.Get("/lessons/{id}/materials", deps.MaterialHandler.List)      // US4.2/US4.3, cần đăng nhập + kiểm tra quyền
			r.Get("/materials/{id}/download", deps.MaterialHandler.Download) // US4.3, cần đăng nhập + kiểm tra quyền

			r.Get("/notifications", deps.NotificationHandler.ListMine)                 // US6.2
			r.Get("/notifications/unread-count", deps.NotificationHandler.UnreadCount) // US6.2
			r.Post("/notifications/{id}/read", deps.NotificationHandler.MarkRead)      // US6.2

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(user.RoleStudent))
				r.Post("/assignments/{id}/submit", deps.SubmissionHandler.Submit)        // US5.2
				r.Get("/assignments/{id}/my-submission", deps.SubmissionHandler.GetMine) // US5.2
				r.Get("/me/progress", deps.ProgressHandler.Me)                           // US7.1
				r.Post("/me/role-upgrade-request", deps.RoleUpgradeHandler.Create)       // US1.7
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(user.RoleTeacher, user.RoleAdmin))
				r.Post("/courses", deps.CourseHandler.Create) // US2.1
				r.Post("/courses/{id}/submit", deps.CourseHandler.SubmitForReview)
				r.Get("/courses/{id}/students", deps.EnrollmentHandler.ListStudents) // US3.3

				r.Post("/courses/{courseId}/chapters", deps.CurriculumHandler.CreateChapter)            // US2.2
				r.Patch("/chapters/{id}", deps.CurriculumHandler.RenameChapter)                         // US4.6
				r.Delete("/chapters/{id}", deps.CurriculumHandler.DeleteChapter)                        // US4.6
				r.Patch("/courses/{courseId}/chapters/reorder", deps.CurriculumHandler.ReorderChapters) // US4.7
				r.Post("/chapters/{chapterId}/lessons", deps.CurriculumHandler.CreateLesson)            // US2.2
				r.Patch("/lessons/{id}", deps.CurriculumHandler.RenameLesson)                           // US4.6
				r.Delete("/lessons/{id}", deps.CurriculumHandler.DeleteLesson)                          // US4.6
				r.Patch("/chapters/{chapterId}/lessons/reorder", deps.CurriculumHandler.ReorderLessons) // US4.7
				r.Post("/lessons/{id}/materials", deps.MaterialHandler.Upload)                          // US4.1
				r.Delete("/materials/{id}", deps.MaterialHandler.Delete)                                // US4.8
				r.Post("/lessons/{id}/assignments", deps.AssignmentHandler.Create)                      // US5.1

				r.Get("/assignments/{id}/submissions", deps.SubmissionHandler.ListByAssignment) // US5.3
				r.Post("/submissions/{id}/grade", deps.SubmissionHandler.Grade)                 // US5.3
				r.Get("/courses/{id}/gradebook", deps.GradebookHandler.ForCourse)               // US5.3

				r.Post("/courses/{id}/notifications", deps.NotificationHandler.SendToCourse) // US6.2

				r.Get("/reports/courses", deps.ReportHandler.Courses) // US7.2
			})

			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.RequireRole(user.RoleAdmin))
				r.Get("/courses/pending", deps.CourseHandler.ListPending)   // US2.3, hàng chờ duyệt
				r.Post("/courses/{id}/approve", deps.CourseHandler.Approve) // US2.3

				r.Get("/role-upgrade-requests", deps.RoleUpgradeHandler.ListPending)           // US1.7
				r.Post("/role-upgrade-requests/{id}/approve", deps.RoleUpgradeHandler.Approve) // US1.7
				r.Post("/role-upgrade-requests/{id}/reject", deps.RoleUpgradeHandler.Reject)   // US1.7
			})
		})
	})

	return r
}
