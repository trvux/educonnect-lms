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

// foreignKeyViolation là mã lỗi Postgres khi 1 thao tác vi phạm ràng buộc
// khóa ngoại (vd DELETE chapter còn lesson tham chiếu tới) — dùng để dịch
// lỗi hạ tầng thành lỗi domain có ý nghĩa nghiệp vụ (US4.6).
const foreignKeyViolation = "23503"

type ChapterRepository struct {
	pool *pgxpool.Pool
}

func NewChapterRepository(pool *pgxpool.Pool) *ChapterRepository {
	return &ChapterRepository{pool: pool}
}

func (r *ChapterRepository) Create(ctx context.Context, c *curriculum.Chapter) error {
	const q = `
		INSERT INTO chapters (course_id, title, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q, c.CourseID(), c.Title(), c.Position(), c.CreatedAt(), c.UpdatedAt()).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo chapter lỗi: %w", err)
	}
	c.SetID(id)
	return nil
}

func (r *ChapterRepository) FindByID(ctx context.Context, id uint) (*curriculum.Chapter, error) {
	const q = `SELECT id, course_id, title, position, created_at, updated_at FROM chapters WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	var (
		chapterID, courseID  uint
		title                string
		position             int
		createdAt, updatedAt time.Time
	)
	if err := row.Scan(&chapterID, &courseID, &title, &position, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, curriculum.ErrChapterNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu chapter lỗi: %w", err)
	}
	return curriculum.RehydrateChapter(chapterID, courseID, title, position, createdAt, updatedAt), nil
}

func (r *ChapterRepository) ListByCourse(ctx context.Context, courseID uint) ([]*curriculum.Chapter, error) {
	const q = `
		SELECT id, course_id, title, position, created_at, updated_at
		FROM chapters WHERE course_id = $1 ORDER BY position ASC`
	rows, err := r.pool.Query(ctx, q, courseID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách chapter lỗi: %w", err)
	}
	defer rows.Close()

	var result []*curriculum.Chapter
	for rows.Next() {
		var (
			id, cID              uint
			title                string
			position             int
			createdAt, updatedAt time.Time
		)
		if err := rows.Scan(&id, &cID, &title, &position, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng chapter lỗi: %w", err)
		}
		result = append(result, curriculum.RehydrateChapter(id, cID, title, position, createdAt, updatedAt))
	}
	return result, rows.Err()
}

func (r *ChapterRepository) CountByCourse(ctx context.Context, courseID uint) (int, error) {
	const q = `SELECT COUNT(*) FROM chapters WHERE course_id = $1`
	var count int
	if err := r.pool.QueryRow(ctx, q, courseID).Scan(&count); err != nil {
		return 0, fmt.Errorf("postgres: đếm chapter lỗi: %w", err)
	}
	return count, nil
}

func (r *ChapterRepository) Update(ctx context.Context, c *curriculum.Chapter) error {
	const q = `UPDATE chapters SET title = $1, position = $2, updated_at = $3 WHERE id = $4`
	tag, err := r.pool.Exec(ctx, q, c.Title(), c.Position(), c.UpdatedAt(), c.ID())
	if err != nil {
		return fmt.Errorf("postgres: cập nhật chapter lỗi: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return curriculum.ErrChapterNotFound
	}
	return nil
}

func (r *ChapterRepository) Delete(ctx context.Context, id uint) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM chapters WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation {
			return curriculum.ErrChapterNotEmpty
		}
		return fmt.Errorf("postgres: xóa chapter lỗi: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return curriculum.ErrChapterNotFound
	}
	return nil
}
