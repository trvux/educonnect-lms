package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	JWTSecret   string
	// AllowRoleOnRegister chỉ nên bật ở môi trường dev/seed dữ liệu demo —
	// cho phép API đăng ký công khai nhận field "role" tuỳ ý (teacher/admin).
	// Khi tắt (mặc định, đúng US1.7), API luôn tạo tài khoản Student, ai
	// muốn làm Giảng viên phải gửi yêu cầu để Admin duyệt.
	AllowRoleOnRegister bool

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

// Load đọc file .env (nếu có — Docker/CI có thể dùng biến môi trường thật
// thay vì .env) và validate các config bắt buộc không được rỗng.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:                 getEnv("APP_ENV", "development"),
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		AllowRoleOnRegister: getEnv("ALLOW_ROLE_ON_REGISTER", "false") == "true",
		SMTPHost:            os.Getenv("SMTP_HOST"),
		SMTPPort:            os.Getenv("SMTP_PORT"),
		SMTPUsername:        os.Getenv("SMTP_USERNAME"),
		// Gmail App Password thường hiển thị có dấu cách (vd "abcd efgh ijkl
		// mnop") để dễ đọc, nhưng SMTP AUTH cần chuỗi liền không dấu cách.
		SMTPPassword: strings.ReplaceAll(os.Getenv("SMTP_PASSWORD"), " ", ""),
		SMTPFrom:     os.Getenv("SMTP_FROM"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("config: DATABASE_URL là bắt buộc")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("config: JWT_SECRET là bắt buộc")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
