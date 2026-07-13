package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/quizattempt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QuizAttemptRepository struct {
	pool *pgxpool.Pool
}

func NewQuizAttemptRepository(pool *pgxpool.Pool) *QuizAttemptRepository {
	return &QuizAttemptRepository{pool: pool}
}

func (r *QuizAttemptRepository) Create(ctx context.Context, a *quizattempt.QuizAttempt) error {
	const q = `
		INSERT INTO quiz_attempts (assignment_id, student_id, started_at)
		VALUES ($1, $2, $3)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q, a.AssignmentID(), a.StudentID(), a.StartedAt()).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			// Race giữa 2 request cùng bắt đầu 1 lúc — attempt đã tồn tại,
			// service sẽ FindByAssignmentAndStudent lại để lấy startedAt gốc.
			return nil
		}
		return fmt.Errorf("postgres: tạo quiz attempt lỗi: %w", err)
	}
	a.SetID(id)
	return nil
}

func (r *QuizAttemptRepository) FindByAssignmentAndStudent(ctx context.Context, assignmentID, studentID uint) (*quizattempt.QuizAttempt, error) {
	const q = `SELECT id, assignment_id, student_id, started_at FROM quiz_attempts WHERE assignment_id = $1 AND student_id = $2`
	var (
		id        uint
		aID, sID  uint
		startedAt time.Time
	)
	err := r.pool.QueryRow(ctx, q, assignmentID, studentID).Scan(&id, &aID, &sID, &startedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, quizattempt.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu quiz attempt lỗi: %w", err)
	}
	return quizattempt.Rehydrate(id, aID, sID, startedAt), nil
}
