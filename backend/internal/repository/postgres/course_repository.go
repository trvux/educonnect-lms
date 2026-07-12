// Package postgres hiện thực các repository port của domain (user.Repository,
// course.Repository) trên nền pgx, viết SQL thuần bám sát schema thay vì
// giấu sau một ORM.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/course"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{pool: pool}
}

func (r *CourseRepository) Create(ctx context.Context, c *course.Course) error {
	const q = `
		INSERT INTO courses (title, description, teacher_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q,
		c.Title(), c.Description(), c.TeacherID(), c.Status(), c.CreatedAt(), c.UpdatedAt(),
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo course lỗi: %w", err)
	}
	c.SetID(id)
	return nil
}

func (r *CourseRepository) FindByID(ctx context.Context, id uint) (*course.Course, error) {
	const q = `
		SELECT id, title, description, teacher_id, status, created_at, updated_at
		FROM courses WHERE id = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, id))
}

// Search hiện thực US3.1: chỉ khóa học đã Approved (US2.3) mới hiện ra
// với học viên, việc tìm kiếm khớp theo title.
func (r *CourseRepository) Search(ctx context.Context, keyword string) ([]*course.Course, error) {
	const q = `
		SELECT id, title, description, teacher_id, status, created_at, updated_at
		FROM courses
		WHERE status = $1 AND title ILIKE '%' || $2 || '%'
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, course.StatusApproved, keyword)
	if err != nil {
		return nil, fmt.Errorf("postgres: tìm kiếm course lỗi: %w", err)
	}
	defer rows.Close()
	return r.scanMany(rows)
}

func (r *CourseRepository) ListByTeacher(ctx context.Context, teacherID uint) ([]*course.Course, error) {
	const q = `
		SELECT id, title, description, teacher_id, status, created_at, updated_at
		FROM courses WHERE teacher_id = $1
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, teacherID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách course theo giáo viên lỗi: %w", err)
	}
	defer rows.Close()
	return r.scanMany(rows)
}

func (r *CourseRepository) Update(ctx context.Context, c *course.Course) error {
	const q = `
		UPDATE courses SET title = $1, description = $2, status = $3, updated_at = $4
		WHERE id = $5`
	tag, err := r.pool.Exec(ctx, q, c.Title(), c.Description(), c.Status(), c.UpdatedAt(), c.ID())
	if err != nil {
		return fmt.Errorf("postgres: cập nhật course lỗi: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return course.ErrNotFound
	}
	return nil
}

func (r *CourseRepository) scanOne(row pgx.Row) (*course.Course, error) {
	var (
		id                  uint
		title, description  string
		teacherID           uint
		status              course.Status
		createdAt, updatedAt time.Time
	)
	err := row.Scan(&id, &title, &description, &teacherID, &status, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, course.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu course lỗi: %w", err)
	}
	return course.Rehydrate(id, title, description, teacherID, status, createdAt, updatedAt), nil
}

func (r *CourseRepository) scanMany(rows pgx.Rows) ([]*course.Course, error) {
	var result []*course.Course
	for rows.Next() {
		var (
			id                  uint
			title, description  string
			teacherID           uint
			status              course.Status
			createdAt, updatedAt time.Time
		)
		if err := rows.Scan(&id, &title, &description, &teacherID, &status, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng course lỗi: %w", err)
		}
		result = append(result, course.Rehydrate(id, title, description, teacherID, status, createdAt, updatedAt))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: duyệt danh sách course lỗi: %w", err)
	}
	return result, nil
}
