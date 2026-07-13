package main

import (
	"context"
	"net/http"
	"time"

	"educonnect-lms/backend/internal/handler"
	"educonnect-lms/backend/internal/platform/config"
	"educonnect-lms/backend/internal/platform/db"
	"educonnect-lms/backend/internal/platform/email"
	"educonnect-lms/backend/internal/platform/logger"
	"educonnect-lms/backend/internal/platform/security"
	"educonnect-lms/backend/internal/platform/storage"
	"educonnect-lms/backend/internal/repository/postgres"
	"educonnect-lms/backend/internal/router"
	assignmentservice "educonnect-lms/backend/internal/service/assignment"
	authservice "educonnect-lms/backend/internal/service/auth"
	courseservice "educonnect-lms/backend/internal/service/course"
	curriculumservice "educonnect-lms/backend/internal/service/curriculum"
	enrollmentservice "educonnect-lms/backend/internal/service/enrollment"
	forumservice "educonnect-lms/backend/internal/service/forum"
	gradebookservice "educonnect-lms/backend/internal/service/gradebook"
	lessoncompletionservice "educonnect-lms/backend/internal/service/lessoncompletion"
	materialservice "educonnect-lms/backend/internal/service/material"
	notificationservice "educonnect-lms/backend/internal/service/notification"
	progressservice "educonnect-lms/backend/internal/service/progress"
	quizattemptservice "educonnect-lms/backend/internal/service/quizattempt"
	reportservice "educonnect-lms/backend/internal/service/report"
	roleupgradeservice "educonnect-lms/backend/internal/service/roleupgrade"
	submissionservice "educonnect-lms/backend/internal/service/submission"

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
	chapterRepo := postgres.NewChapterRepository(pool)
	lessonRepo := postgres.NewLessonRepository(pool)
	enrollmentRepo := postgres.NewEnrollmentRepository(pool)
	materialRepo := postgres.NewMaterialRepository(pool)
	assignmentRepo := postgres.NewAssignmentRepository(pool)
	submissionRepo := postgres.NewSubmissionRepository(pool)
	gradebookRepo := postgres.NewGradebookRepository(pool)
	forumRepo := postgres.NewForumRepository(pool)
	notificationRepo := postgres.NewNotificationRepository(pool)
	progressRepo := postgres.NewProgressRepository(pool)
	reportRepo := postgres.NewReportRepository(pool)
	passwordResetRepo := postgres.NewPasswordResetRepository(pool)
	emailVerificationRepo := postgres.NewEmailVerificationRepository(pool)
	roleUpgradeRepo := postgres.NewRoleUpgradeRepository(pool)
	lessonCompletionRepo := postgres.NewLessonCompletionRepository(pool)
	quizAttemptRepo := postgres.NewQuizAttemptRepository(pool)
	hasher := security.NewBcryptHasher()
	tokens := security.NewJWTIssuer(cfg.JWTSecret, 24*time.Hour)
	// US4.5: token riêng cho <video src="...">, TTL ngắn hơn hẳn JWT đăng
	// nhập chính vì token này lộ ra trong query string URL (xem thiết kế ở
	// Câu 3b báo cáo Sprint 5).
	streamTokens := security.NewJWTIssuer(cfg.JWTSecret, 30*time.Minute)
	fileStorage, err := storage.NewLocalFileStorage("uploads")
	if err != nil {
		log.Fatal("khởi tạo file storage thất bại", zap.Error(err))
	}
	emailSender := email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)

	// tầng service (application)
	authSvc := authservice.NewService(userRepo, hasher, tokens, fileStorage, passwordResetRepo, emailVerificationRepo, emailSender)
	roleUpgradeSvc := roleupgradeservice.NewService(roleUpgradeRepo, userRepo)
	courseSvc := courseservice.NewService(courseRepo)
	curriculumSvc := curriculumservice.NewService(chapterRepo, lessonRepo, courseRepo)
	enrollmentSvc := enrollmentservice.NewService(enrollmentRepo, courseRepo, userRepo)
	materialSvc := materialservice.NewService(materialRepo, lessonRepo, chapterRepo, courseRepo, enrollmentRepo, fileStorage)
	assignmentSvc := assignmentservice.NewService(assignmentRepo, lessonRepo)
	submissionSvc := submissionservice.NewService(submissionRepo, assignmentSvc, lessonRepo, chapterRepo, courseRepo, quizAttemptRepo)
	gradebookSvc := gradebookservice.NewService(gradebookRepo)
	forumSvc := forumservice.NewService(forumRepo, courseRepo)
	notificationSvc := notificationservice.NewService(notificationRepo, enrollmentRepo, courseRepo)
	progressSvc := progressservice.NewService(progressRepo)
	reportSvc := reportservice.NewService(reportRepo)
	lessonCompletionSvc := lessoncompletionservice.NewService(lessonCompletionRepo, lessonRepo, chapterRepo, courseRepo, enrollmentRepo)
	quizAttemptSvc := quizattemptservice.NewService(quizAttemptRepo, assignmentSvc)

	// tầng HTTP
	authHandler := handler.NewAuthHandler(authSvc, log, cfg.AllowRoleOnRegister)
	courseHandler := handler.NewCourseHandler(courseSvc, log)
	curriculumHandler := handler.NewCurriculumHandler(curriculumSvc, log)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentSvc, log)
	materialHandler := handler.NewMaterialHandler(materialSvc, log, "uploads", streamTokens)
	assignmentHandler := handler.NewAssignmentHandler(assignmentSvc, log)
	submissionHandler := handler.NewSubmissionHandler(submissionSvc, log)
	gradebookHandler := handler.NewGradebookHandler(gradebookSvc, courseSvc, log)
	forumHandler := handler.NewForumHandler(forumSvc, log)
	notificationHandler := handler.NewNotificationHandler(notificationSvc, log)
	progressHandler := handler.NewProgressHandler(progressSvc, log)
	reportHandler := handler.NewReportHandler(reportSvc, log)
	passwordResetHandler := handler.NewPasswordResetHandler(authSvc, log)
	roleUpgradeHandler := handler.NewRoleUpgradeHandler(roleUpgradeSvc, log)
	lessonCompletionHandler := handler.NewLessonCompletionHandler(lessonCompletionSvc, log)
	quizAttemptHandler := handler.NewQuizAttemptHandler(quizAttemptSvc, log)

	r := router.New(router.Deps{
		AuthHandler:             authHandler,
		CourseHandler:           courseHandler,
		CurriculumHandler:       curriculumHandler,
		EnrollmentHandler:       enrollmentHandler,
		MaterialHandler:         materialHandler,
		AssignmentHandler:       assignmentHandler,
		SubmissionHandler:       submissionHandler,
		GradebookHandler:        gradebookHandler,
		ForumHandler:            forumHandler,
		NotificationHandler:     notificationHandler,
		ProgressHandler:         progressHandler,
		ReportHandler:           reportHandler,
		PasswordResetHandler:    passwordResetHandler,
		RoleUpgradeHandler:      roleUpgradeHandler,
		LessonCompletionHandler: lessonCompletionHandler,
		QuizAttemptHandler:      quizAttemptHandler,
		TokenVerifier:           tokens,
		StreamTokenVerifier:     streamTokens,
		UploadsDir:              "uploads",
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
