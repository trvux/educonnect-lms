package logger

import "go.uber.org/zap"

// New dựng logger structured dùng chung toàn process. Bản dev in ra console
// dễ đọc; set APP_ENV=production để log dạng JSON phù hợp cho log aggregator.
func New(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
