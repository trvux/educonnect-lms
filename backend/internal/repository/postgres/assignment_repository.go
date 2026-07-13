package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/assignment"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AssignmentRepository struct {
	pool *pgxpool.Pool
}

func NewAssignmentRepository(pool *pgxpool.Pool) *AssignmentRepository {
	return &AssignmentRepository{pool: pool}
}

// questionDTO là dạng trung gian để (de)serialize assignment.Question sang
// cột JSONB — domain entity không phụ thuộc trực tiếp encoding/json.
type questionDTO struct {
	Content      string   `json:"content"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
}

func encodeQuestions(qs []assignment.Question) ([]byte, error) {
	dto := make([]questionDTO, len(qs))
	for i, q := range qs {
		dto[i] = questionDTO{Content: q.Content, Options: q.Options, CorrectIndex: q.CorrectIndex}
	}
	return json.Marshal(dto)
}

func decodeQuestions(data []byte) ([]assignment.Question, error) {
	var dto []questionDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}
	if len(dto) == 0 {
		return nil, nil
	}
	out := make([]assignment.Question, len(dto))
	for i, q := range dto {
		out[i] = assignment.Question{Content: q.Content, Options: q.Options, CorrectIndex: q.CorrectIndex}
	}
	return out, nil
}

func (r *AssignmentRepository) Create(ctx context.Context, a *assignment.Assignment) error {
	questionsData, err := encodeQuestions(a.Questions())
	if err != nil {
		return fmt.Errorf("postgres: encode câu hỏi lỗi: %w", err)
	}

	const q = `
		INSERT INTO assignments (lesson_id, title, description, kind, questions, due_at, time_limit_minutes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	var id uint
	err = r.pool.QueryRow(ctx, q,
		a.LessonID(), a.Title(), a.Description(), string(a.Kind()), questionsData, a.DueAt(), a.TimeLimitMinutes(), a.CreatedAt(),
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo assignment lỗi: %w", err)
	}
	a.SetID(id)
	return nil
}

func (r *AssignmentRepository) FindByID(ctx context.Context, id uint) (*assignment.Assignment, error) {
	const q = `SELECT id, lesson_id, title, description, kind, questions, due_at, time_limit_minutes, created_at FROM assignments WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	a, err := scanAssignment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, assignment.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu assignment lỗi: %w", err)
	}
	return a, nil
}

func (r *AssignmentRepository) ListByLesson(ctx context.Context, lessonID uint) ([]*assignment.Assignment, error) {
	const q = `
		SELECT id, lesson_id, title, description, kind, questions, due_at, time_limit_minutes, created_at
		FROM assignments WHERE lesson_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, lessonID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách assignment lỗi: %w", err)
	}
	defer rows.Close()

	var result []*assignment.Assignment
	for rows.Next() {
		a, err := scanAssignment(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng assignment lỗi: %w", err)
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// rowScanner khớp cả pgx.Row (QueryRow) lẫn pgx.Rows (Query), tránh lặp code
// giữa FindByID và ListByLesson.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanAssignment(row rowScanner) (*assignment.Assignment, error) {
	var (
		id, lessonID       uint
		title, description string
		kind               string
		questionsData      []byte
		dueAt              *time.Time
		timeLimitMinutes   *int
		createdAt          time.Time
	)
	if err := row.Scan(&id, &lessonID, &title, &description, &kind, &questionsData, &dueAt, &timeLimitMinutes, &createdAt); err != nil {
		return nil, err
	}
	questions, err := decodeQuestions(questionsData)
	if err != nil {
		return nil, fmt.Errorf("decode câu hỏi lỗi: %w", err)
	}
	return assignment.Rehydrate(id, lessonID, title, description, assignment.Type(kind), questions, dueAt, timeLimitMinutes, createdAt), nil
}
