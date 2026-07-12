package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// openTestPool kết nối tới Postgres đã khởi động bằng `make up` (+ `make
// migrate-up`). Test repository về bản chất là integration test (đúng mục
// đích: chứng minh SQL đúng trên database thật), nên sẽ skip thay vì fail
// khi DATABASE_URL chưa được set, ví dụ ở môi trường chỉ chạy unit test nhanh.
func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set — run `make up && make migrate-up` then re-run tests")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)
	truncateAll(t, pool)
	return pool
}

// truncateAll reset bảng giữa các lần chạy test để mỗi test bắt đầu từ
// trạng thái sạch, không cần chạy lại cả chu trình migrate-down/up.
func truncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), "TRUNCATE enrollments, courses, users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}
