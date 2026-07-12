package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool mở connection pool pgx và kiểm tra kết nối bằng ping, để main.go
// fail ngay từ đầu (thay vì fail ở request đầu tiên) nếu Postgres bị down.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("db: tạo connection pool lỗi: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping database lỗi: %w", err)
	}
	return pool, nil
}
