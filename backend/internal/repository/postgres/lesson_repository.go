package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/curriculum"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LessonRepository struct {
	pool *pgxpool.Pool
}

func NewLessonRepository(pool *pgxpool.Pool) *LessonRepository {
	return &LessonRepository{pool: pool}
}

func (r *LessonRepository) Create(ctx context.Context, l *curriculum.Lesson) error {
	const q = `
		INSERT INTO lessons (chapter_id, title, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q, l.ChapterID(), l.Title(), l.Position(), l.CreatedAt(), l.UpdatedAt()).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo lesson lỗi: %w", err)
	}
	l.SetID(id)
	return nil
}

func (r *LessonRepository) FindByID(ctx context.Context, id uint) (*curriculum.Lesson, error) {
	const q = `SELECT id, chapter_id, title, position, created_at, updated_at FROM lessons WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	var (
		lessonID, chapterID  uint
		title                string
		position             int
		createdAt, updatedAt time.Time
	)
	if err := row.Scan(&lessonID, &chapterID, &title, &position, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, curriculum.ErrLessonNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu lesson lỗi: %w", err)
	}
	return curriculum.RehydrateLesson(lessonID, chapterID, title, position, createdAt, updatedAt), nil
}

func (r *LessonRepository) ListByChapter(ctx context.Context, chapterID uint) ([]*curriculum.Lesson, error) {
	const q = `
		SELECT id, chapter_id, title, position, created_at, updated_at
		FROM lessons WHERE chapter_id = $1 ORDER BY position ASC`
	rows, err := r.pool.Query(ctx, q, chapterID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách lesson lỗi: %w", err)
	}
	defer rows.Close()

	var result []*curriculum.Lesson
	for rows.Next() {
		var (
			id, cID              uint
			title                string
			position             int
			createdAt, updatedAt time.Time
		)
		if err := rows.Scan(&id, &cID, &title, &position, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng lesson lỗi: %w", err)
		}
		result = append(result, curriculum.RehydrateLesson(id, cID, title, position, createdAt, updatedAt))
	}
	return result, rows.Err()
}

func (r *LessonRepository) CountByChapter(ctx context.Context, chapterID uint) (int, error) {
	const q = `SELECT COUNT(*) FROM lessons WHERE chapter_id = $1`
	var count int
	if err := r.pool.QueryRow(ctx, q, chapterID).Scan(&count); err != nil {
		return 0, fmt.Errorf("postgres: đếm lesson lỗi: %w", err)
	}
	return count, nil
}

func (r *LessonRepository) Update(ctx context.Context, l *curriculum.Lesson) error {
	const q = `UPDATE lessons SET title = $1, position = $2, updated_at = $3 WHERE id = $4`
	tag, err := r.pool.Exec(ctx, q, l.Title(), l.Position(), l.UpdatedAt(), l.ID())
	if err != nil {
		return fmt.Errorf("postgres: cập nhật lesson lỗi: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return curriculum.ErrLessonNotFound
	}
	return nil
}

func (r *LessonRepository) Delete(ctx context.Context, id uint) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM lessons WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation {
			return curriculum.ErrLessonNotEmpty
		}
		return fmt.Errorf("postgres: xóa lesson lỗi: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return curriculum.ErrLessonNotFound
	}
	return nil
}
