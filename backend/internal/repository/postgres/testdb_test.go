package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// openTestPool connects to the Postgres started by `make up` (+ `make
// migrate-up`). Repository tests are integration tests by nature (that's the
// point: prove the SQL is correct against a real database) so they skip
// instead of failing when DATABASE_URL isn't set, e.g. in an environment
// that only runs the fast unit tests.
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

// truncateAll resets tables between test runs so each test starts from a
// clean slate without needing a full migrate-down/up cycle.
func truncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), "TRUNCATE enrollments, courses, users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}
