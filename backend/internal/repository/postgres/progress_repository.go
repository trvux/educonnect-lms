package postgres

import (
	"context"
	"fmt"

	"educonnect-lms/backend/internal/domain/progress"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProgressRepository struct {
	pool *pgxpool.Pool
}

func NewProgressRepository(pool *pgxpool.Pool) *ProgressRepository {
	return &ProgressRepository{pool: pool}
}

func (r *ProgressRepository) ForStudent(ctx context.Context, studentID uint) ([]progress.CourseProgress, error) {
	const q = `
		SELECT
			c.id,
			c.title,
			count(DISTINCT a.id) AS total_assignments,
			count(DISTINCT s.id) AS submitted
		FROM enrollments e
		JOIN courses c ON c.id = e.course_id
		LEFT JOIN chapters ch ON ch.course_id = c.id
		LEFT JOIN lessons l ON l.chapter_id = ch.id
		LEFT JOIN assignments a ON a.lesson_id = l.id
		LEFT JOIN submissions s ON s.assignment_id = a.id AND s.student_id = e.student_id
		WHERE e.student_id = $1
		GROUP BY c.id, c.title
		ORDER BY c.title`
	rows, err := r.pool.Query(ctx, q, studentID)
	if err != nil {
		return nil, fmt.Errorf("postgres: truy vấn tiến độ học viên lỗi: %w", err)
	}
	defer rows.Close()

	var result []progress.CourseProgress
	for rows.Next() {
		var p progress.CourseProgress
		if err := rows.Scan(&p.CourseID, &p.CourseTitle, &p.TotalAssignments, &p.Submitted); err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng tiến độ học viên lỗi: %w", err)
		}
		if p.TotalAssignments > 0 {
			p.PercentComplete = float64(p.Submitted) / float64(p.TotalAssignments) * 100
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
