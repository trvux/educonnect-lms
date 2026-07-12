package main

import (
	"context"
	"net/http"
	"time"

	"educonnect-lms/backend/internal/handler"
	"educonnect-lms/backend/internal/platform/config"
	"educonnect-lms/backend/internal/platform/db"
	"educonnect-lms/backend/internal/platform/logger"
	"educonnect-lms/backend/internal/platform/security"
	"educonnect-lms/backend/internal/repository/postgres"
	"educonnect-lms/backend/internal/router"
	authservice "educonnect-lms/backend/internal/service/auth"
	courseservice "educonnect-lms/backend/internal/service/course"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Env)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("kết nối database thất bại", zap.Error(err))
	}
	defer pool.Close()

	// tầng infrastructure
	userRepo := postgres.NewUserRepository(pool)
	courseRepo := postgres.NewCourseRepository(pool)
	hasher := security.NewBcryptHasher()
	tokens := security.NewJWTIssuer(cfg.JWTSecret, 24*time.Hour)

	// tầng service (application)
	authSvc := authservice.NewService(userRepo, hasher, tokens)
	courseSvc := courseservice.NewService(courseRepo)

	// tầng HTTP
	authHandler := handler.NewAuthHandler(authSvc, log)
	courseHandler := handler.NewCourseHandler(courseSvc, log)

	r := router.New(router.Deps{
		AuthHandler:   authHandler,
		CourseHandler: courseHandler,
		TokenVerifier: tokens,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Info("khởi động EduConnect LMS API", zap.String("port", cfg.Port), zap.String("env", cfg.Env))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("lỗi server", zap.Error(err))
	}
}
