package postgres

import (
	"context"
	"fmt"

	"educonnect-lms/backend/internal/domain/report"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	pool *pgxpool.Pool
}

func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

const courseStatsQuery = `
	SELECT
		c.id,
		c.title,
		count(DISTINCT e.student_id) AS enrolled_students,
		count(DISTINCT a.id) AS total_assignments,
		count(DISTINCT s.id) AS total_submissions
	FROM courses c
	LEFT JOIN enrollments e ON e.course_id = c.id
	LEFT JOIN chapters ch ON ch.course_id = c.id
	LEFT JOIN lessons l ON l.chapter_id = ch.id
	LEFT JOIN assignments a ON a.lesson_id = l.id
	LEFT JOIN submissions s ON s.assignment_id = a.id AND s.student_id = e.student_id`

func (r *ReportRepository) ForTeacher(ctx context.Context, teacherID uint) ([]report.CourseStats, error) {
	q := courseStatsQuery + ` WHERE c.teacher_id = $1 GROUP BY c.id, c.title ORDER BY c.title`
	return r.queryCourseStats(ctx, q, teacherID)
}

func (r *ReportRepository) All(ctx context.Context) ([]report.CourseStats, error) {
	q := courseStatsQuery + ` GROUP BY c.id, c.title ORDER BY c.title`
	return r.queryCourseStats(ctx, q)
}

func (r *ReportRepository) queryCourseStats(ctx context.Context, q string, args ...any) ([]report.CourseStats, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: truy vấn báo cáo khóa học lỗi: %w", err)
	}
	defer rows.Close()

	var result []report.CourseStats
	for rows.Next() {
		var (
			s                report.CourseStats
			totalSubmissions int
		)
		if err := rows.Scan(&s.CourseID, &s.CourseTitle, &s.EnrolledStudents, &s.TotalAssignments, &totalSubmissions); err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng báo cáo khóa học lỗi: %w", err)
		}
		if denom := s.EnrolledStudents * s.TotalAssignments; denom > 0 {
			s.AverageCompletion = float64(totalSubmissions) / float64(denom) * 100
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
