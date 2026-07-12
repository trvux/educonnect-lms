package router

import (
	"net/http"
	"time"

	"educonnect-lms/backend/internal/handler"
	"educonnect-lms/backend/internal/handler/middleware"
	"educonnect-lms/backend/internal/domain/user"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Deps struct {
	AuthHandler   *handler.AuthHandler
	CourseHandler *handler.CourseHandler
	TokenVerifier middleware.TokenVerifier
}

func New(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", deps.AuthHandler.Register) // US1.1
			r.Post("/login", deps.AuthHandler.Login)        // US1.2
		})

		r.Get("/courses", deps.CourseHandler.Search) // US3.1, public

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(deps.TokenVerifier))

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(user.RoleTeacher, user.RoleAdmin))
				r.Post("/courses", deps.CourseHandler.Create)                  // US2.1
				r.Post("/courses/{id}/submit", deps.CourseHandler.SubmitForReview)
			})

			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.RequireRole(user.RoleAdmin))
				r.Post("/courses/{id}/approve", deps.CourseHandler.Approve) // US2.3
			})
		})
	})

	return r
}
