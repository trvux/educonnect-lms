package postgres

import (
	"context"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/enrollment"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EnrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{pool: pool}
}

func (r *EnrollmentRepository) Create(ctx context.Context, e *enrollment.Enrollment) error {
	const q = `
		INSERT INTO enrollments (student_id, course_id, enrolled_at)
		VALUES ($1, $2, $3)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q, e.StudentID(), e.CourseID(), e.EnrolledAt()).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo enrollment lỗi: %w", err)
	}
	e.SetID(id)
	return nil
}

// IsEnrolled dùng ở service để chặn đăng ký trùng (US3.2) trước khi Create,
// dựa trên UNIQUE(student_id, course_id) đã khai báo ở migration.
func (r *EnrollmentRepository) IsEnrolled(ctx context.Context, studentID, courseID uint) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM enrollments WHERE student_id = $1 AND course_id = $2)`
	var exists bool
	if err := r.pool.QueryRow(ctx, q, studentID, courseID).Scan(&exists); err != nil {
		return false, fmt.Errorf("postgres: kiểm tra enrollment lỗi: %w", err)
	}
	return exists, nil
}

func (r *EnrollmentRepository) ListByCourse(ctx context.Context, courseID uint) ([]*enrollment.Enrollment, error) {
	const q = `SELECT id, student_id, course_id, enrolled_at FROM enrollments WHERE course_id = $1 ORDER BY enrolled_at ASC`
	rows, err := r.pool.Query(ctx, q, courseID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách enrollment theo course lỗi: %w", err)
	}
	defer rows.Close()
	return scanEnrollments(rows)
}

func (r *EnrollmentRepository) ListByStudent(ctx context.Context, studentID uint) ([]*enrollment.Enrollment, error) {
	const q = `SELECT id, student_id, course_id, enrolled_at FROM enrollments WHERE student_id = $1 ORDER BY enrolled_at ASC`
	rows, err := r.pool.Query(ctx, q, studentID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách enrollment theo học viên lỗi: %w", err)
	}
	defer rows.Close()
	return scanEnrollments(rows)
}

func scanEnrollments(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]*enrollment.Enrollment, error) {
	var result []*enrollment.Enrollment
	for rows.Next() {
		var (
			id, studentID, courseID uint
			enrolledAt              time.Time
		)
		if err := rows.Scan(&id, &studentID, &courseID, &enrolledAt); err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng enrollment lỗi: %w", err)
		}
		result = append(result, enrollment.Rehydrate(id, studentID, courseID, enrolledAt))
	}
	return result, rows.Err()
}
