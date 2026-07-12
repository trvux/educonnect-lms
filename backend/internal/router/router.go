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
	AuthHandler       *handler.AuthHandler
	CourseHandler     *handler.CourseHandler
	CurriculumHandler *handler.CurriculumHandler
	EnrollmentHandler *handler.EnrollmentHandler
	MaterialHandler   *handler.MaterialHandler
	TokenVerifier     middleware.TokenVerifier
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

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", deps.AuthHandler.Register) // US1.1
			r.Post("/login", deps.AuthHandler.Login)       // US1.2
		})

		r.Get("/courses", deps.CourseHandler.Search)                               // US3.1, public
		r.Get("/courses/{courseId}/chapters", deps.CurriculumHandler.ListChapters) // US2.2, public
		r.Get("/chapters/{chapterId}/lessons", deps.CurriculumHandler.ListLessons) // US2.2, public
		r.Get("/lessons/{id}/materials", deps.MaterialHandler.List)                // US4.2, public

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(deps.TokenVerifier))

			r.Post("/courses/{id}/enroll", deps.EnrollmentHandler.Enroll) // US3.2, mọi user đã đăng nhập

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(user.RoleTeacher, user.RoleAdmin))
				r.Post("/courses", deps.CourseHandler.Create) // US2.1
				r.Post("/courses/{id}/submit", deps.CourseHandler.SubmitForReview)
				r.Get("/courses/{id}/students", deps.EnrollmentHandler.ListStudents) // US3.3

				r.Post("/courses/{courseId}/chapters", deps.CurriculumHandler.CreateChapter) // US2.2
				r.Post("/chapters/{chapterId}/lessons", deps.CurriculumHandler.CreateLesson) // US2.2
				r.Post("/lessons/{id}/materials", deps.MaterialHandler.Upload)               // US4.1
			})

			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.RequireRole(user.RoleAdmin))
				r.Post("/courses/{id}/approve", deps.CourseHandler.Approve) // US2.3
			})
		})
	})

	return r
}
