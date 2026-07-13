package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/material"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MaterialRepository struct {
	pool *pgxpool.Pool
}

func NewMaterialRepository(pool *pgxpool.Pool) *MaterialRepository {
	return &MaterialRepository{pool: pool}
}

func (r *MaterialRepository) Create(ctx context.Context, m *material.Material) error {
	const q = `
		INSERT INTO materials (lesson_id, file_name, file_path, file_type, uploaded_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q, m.LessonID(), m.FileName(), m.FilePath(), string(m.FileType()), m.UploadedAt()).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo material lỗi: %w", err)
	}
	m.SetID(id)
	return nil
}

func (r *MaterialRepository) FindByID(ctx context.Context, id uint) (*material.Material, error) {
	const q = `SELECT id, lesson_id, file_name, file_path, file_type, uploaded_at FROM materials WHERE id = $1`
	var (
		mID, lID           uint
		fileName, filePath string
		fileType           string
		uploadedAt         time.Time
	)
	err := r.pool.QueryRow(ctx, q, id).Scan(&mID, &lID, &fileName, &filePath, &fileType, &uploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, material.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu material lỗi: %w", err)
	}
	return material.Rehydrate(mID, lID, fileName, filePath, material.FileType(fileType), uploadedAt), nil
}

func (r *MaterialRepository) ListByLesson(ctx context.Context, lessonID uint) ([]*material.Material, error) {
	const q = `SELECT id, lesson_id, file_name, file_path, file_type, uploaded_at FROM materials WHERE lesson_id = $1 ORDER BY uploaded_at ASC`
	rows, err := r.pool.Query(ctx, q, lessonID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách material lỗi: %w", err)
	}
	defer rows.Close()

	var result []*material.Material
	for rows.Next() {
		var (
			id, lID            uint
			fileName, filePath string
			fileType           string
			uploadedAt         time.Time
		)
		if err := rows.Scan(&id, &lID, &fileName, &filePath, &fileType, &uploadedAt); err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng material lỗi: %w", err)
		}
		result = append(result, material.Rehydrate(id, lID, fileName, filePath, material.FileType(fileType), uploadedAt))
	}
	return result, rows.Err()
}
