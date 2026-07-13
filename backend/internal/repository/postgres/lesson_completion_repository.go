package postgres

import (
	"context"
	"errors"
	"fmt"

	"educonnect-lms/backend/internal/domain/lessoncompletion"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// uniqueViolation là mã lỗi Postgres khi vi phạm ràng buộc UNIQUE — ở đây
// dùng để coi việc đánh dấu hoàn thành trùng lặp (race giữa 2 request) là
// thành công thay vì lỗi (US4.10, idempotent).
const uniqueViolation = "23505"

type LessonCompletionRepository struct {
	pool *pgxpool.Pool
}

func NewLessonCompletionRepository(pool *pgxpool.Pool) *LessonCompletionRepository {
	return &LessonCompletionRepository{pool: pool}
}

func (r *LessonCompletionRepository) Create(ctx context.Context, c *lessoncompletion.LessonCompletion) error {
	const q = `
		INSERT INTO lesson_completions (student_id, lesson_id, completed_at)
		VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, q, c.StudentID(), c.LessonID(), c.CompletedAt())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return nil
		}
		return fmt.Errorf("postgres: tạo lesson completion lỗi: %w", err)
	}
	return nil
}

func (r *LessonCompletionRepository) IsCompleted(ctx context.Context, studentID, lessonID uint) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM lesson_completions WHERE student_id = $1 AND lesson_id = $2)`
	var exists bool
	if err := r.pool.QueryRow(ctx, q, studentID, lessonID).Scan(&exists); err != nil {
		return false, fmt.Errorf("postgres: kiểm tra lesson completion lỗi: %w", err)
	}
	return exists, nil
}

func (r *LessonCompletionRepository) ListCompletedByStudent(ctx context.Context, studentID uint) (map[uint]bool, error) {
	const q = `SELECT lesson_id FROM lesson_completions WHERE student_id = $1`
	rows, err := r.pool.Query(ctx, q, studentID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách lesson completion lỗi: %w", err)
	}
	defer rows.Close()

	out := map[uint]bool{}
	for rows.Next() {
		var lessonID uint
		if err := rows.Scan(&lessonID); err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng lesson completion lỗi: %w", err)
		}
		out[lessonID] = true
	}
	return out, rows.Err()
}
