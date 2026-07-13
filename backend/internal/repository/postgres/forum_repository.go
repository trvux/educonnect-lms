package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/forum"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ForumRepository struct {
	pool *pgxpool.Pool
}

func NewForumRepository(pool *pgxpool.Pool) *ForumRepository {
	return &ForumRepository{pool: pool}
}

func (r *ForumRepository) Create(ctx context.Context, p *forum.Post) error {
	const q = `
		INSERT INTO forum_posts (course_id, author_id, parent_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q, p.CourseID(), p.AuthorID(), p.ParentID(), p.Content(), p.CreatedAt()).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo forum post lỗi: %w", err)
	}
	p.SetID(id)
	return nil
}

func (r *ForumRepository) FindByID(ctx context.Context, id uint) (*forum.Post, error) {
	const q = `SELECT id, course_id, author_id, parent_id, content, created_at FROM forum_posts WHERE id = $1`
	var (
		postID, courseID, authorID uint
		parentID                   *uint
		content                    string
		createdAt                  time.Time
	)
	err := r.pool.QueryRow(ctx, q, id).Scan(&postID, &courseID, &authorID, &parentID, &content, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, forum.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu forum post lỗi: %w", err)
	}
	return forum.Rehydrate(postID, courseID, authorID, parentID, content, createdAt), nil
}

func (r *ForumRepository) ListByCourse(ctx context.Context, courseID uint) ([]*forum.Post, error) {
	const q = `
		SELECT p.id, p.course_id, p.author_id, p.parent_id, p.content, p.created_at, u.full_name
		FROM forum_posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.course_id = $1
		ORDER BY p.created_at ASC`
	rows, err := r.pool.Query(ctx, q, courseID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách forum post lỗi: %w", err)
	}
	defer rows.Close()

	var result []*forum.Post
	for rows.Next() {
		var (
			id, cID, authorID uint
			parentID          *uint
			content, author   string
			createdAt         time.Time
		)
		if err := rows.Scan(&id, &cID, &authorID, &parentID, &content, &createdAt, &author); err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng forum post lỗi: %w", err)
		}
		result = append(result, forum.RehydrateWithAuthor(id, cID, authorID, parentID, content, createdAt, author))
	}
	return result, rows.Err()
}
