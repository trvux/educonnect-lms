package logger

import "go.uber.org/zap"

// New builds the process-wide structured logger. Dev builds get readable
// console output; set APP_ENV=production for JSON logs suited to log
// aggregators.
func New(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
