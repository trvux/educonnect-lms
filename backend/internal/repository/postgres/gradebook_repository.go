package postgres

import (
	"context"
	"fmt"

	"educonnect-lms/backend/internal/domain/gradebook"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GradebookRepository struct {
	pool *pgxpool.Pool
}

func NewGradebookRepository(pool *pgxpool.Pool) *GradebookRepository {
	return &GradebookRepository{pool: pool}
}

func (r *GradebookRepository) ForCourse(ctx context.Context, courseID uint) ([]gradebook.Entry, error) {
	const q = `
		SELECT u.id, u.full_name, a.id, a.title, s.score
		FROM enrollments e
		JOIN users u ON u.id = e.student_id
		JOIN chapters c ON c.course_id = e.course_id
		JOIN lessons l ON l.chapter_id = c.id
		JOIN assignments a ON a.lesson_id = l.id
		LEFT JOIN submissions s ON s.assignment_id = a.id AND s.student_id = u.id
		WHERE e.course_id = $1
		ORDER BY u.full_name, a.id`
	rows, err := r.pool.Query(ctx, q, courseID)
	if err != nil {
		return nil, fmt.Errorf("postgres: truy vấn bảng điểm lỗi: %w", err)
	}
	defer rows.Close()

	var result []gradebook.Entry
	for rows.Next() {
		var e gradebook.Entry
		if err := rows.Scan(&e.StudentID, &e.StudentName, &e.AssignmentID, &e.AssignmentTitle, &e.Score); err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng bảng điểm lỗi: %w", err)
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
