package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/submission"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubmissionRepository struct {
	pool *pgxpool.Pool
}

func NewSubmissionRepository(pool *pgxpool.Pool) *SubmissionRepository {
	return &SubmissionRepository{pool: pool}
}

func (r *SubmissionRepository) Create(ctx context.Context, s *submission.Submission) error {
	answersData, err := json.Marshal(s.Answers())
	if err != nil {
		return fmt.Errorf("postgres: encode đáp án lỗi: %w", err)
	}

	const q = `
		INSERT INTO submissions (assignment_id, student_id, content, answers, submitted_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	var id uint
	err = r.pool.QueryRow(ctx, q, s.AssignmentID(), s.StudentID(), s.Content(), answersData, s.SubmittedAt()).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo submission lỗi: %w", err)
	}
	s.SetID(id)
	return nil
}

func (r *SubmissionRepository) FindByAssignmentAndStudent(ctx context.Context, assignmentID, studentID uint) (*submission.Submission, error) {
	const q = `
		SELECT id, assignment_id, student_id, content, answers, submitted_at
		FROM submissions WHERE assignment_id = $1 AND student_id = $2`
	row := r.pool.QueryRow(ctx, q, assignmentID, studentID)
	s, err := scanSubmission(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, submission.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu submission lỗi: %w", err)
	}
	return s, nil
}

func (r *SubmissionRepository) ListByAssignment(ctx context.Context, assignmentID uint) ([]*submission.Submission, error) {
	const q = `
		SELECT id, assignment_id, student_id, content, answers, submitted_at
		FROM submissions WHERE assignment_id = $1 ORDER BY submitted_at ASC`
	rows, err := r.pool.Query(ctx, q, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách submission lỗi: %w", err)
	}
	defer rows.Close()

	var result []*submission.Submission
	for rows.Next() {
		s, err := scanSubmission(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng submission lỗi: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func scanSubmission(row rowScanner) (*submission.Submission, error) {
	var (
		id, assignmentID, studentID uint
		content                     string
		answersData                 []byte
		submittedAt                 time.Time
	)
	if err := row.Scan(&id, &assignmentID, &studentID, &content, &answersData, &submittedAt); err != nil {
		return nil, err
	}
	var answers []int
	if err := json.Unmarshal(answersData, &answers); err != nil {
		return nil, fmt.Errorf("decode đáp án lỗi: %w", err)
	}
	return submission.Rehydrate(id, assignmentID, studentID, content, answers, submittedAt), nil
}
