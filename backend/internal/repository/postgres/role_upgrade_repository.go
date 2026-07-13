package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/roleupgrade"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleUpgradeRepository struct {
	pool *pgxpool.Pool
}

func NewRoleUpgradeRepository(pool *pgxpool.Pool) *RoleUpgradeRepository {
	return &RoleUpgradeRepository{pool: pool}
}

func (r *RoleUpgradeRepository) Create(ctx context.Context, req *roleupgrade.Request) error {
	const q = `
		INSERT INTO role_upgrade_requests (user_id, reason, status, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q, req.UserID(), req.Reason(), string(req.Status()), req.CreatedAt()).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo role upgrade request lỗi: %w", err)
	}
	req.SetID(id)
	return nil
}

func (r *RoleUpgradeRepository) FindByID(ctx context.Context, id uint) (*roleupgrade.Request, error) {
	const q = `
		SELECT id, user_id, reason, status, reviewed_by, created_at, reviewed_at
		FROM role_upgrade_requests WHERE id = $1`
	return scanRoleUpgrade(r.pool.QueryRow(ctx, q, id))
}

// FindPendingByUser trả về ErrNotFound nếu user chưa có yêu cầu đang chờ —
// dùng để chặn gửi trùng (US1.7, có index UNIQUE hỗ trợ ở tầng DB).
func (r *RoleUpgradeRepository) FindPendingByUser(ctx context.Context, userID uint) (*roleupgrade.Request, error) {
	const q = `
		SELECT id, user_id, reason, status, reviewed_by, created_at, reviewed_at
		FROM role_upgrade_requests WHERE user_id = $1 AND status = 'pending'`
	return scanRoleUpgrade(r.pool.QueryRow(ctx, q, userID))
}

func (r *RoleUpgradeRepository) ListPending(ctx context.Context) ([]*roleupgrade.Request, error) {
	const q = `
		SELECT id, user_id, reason, status, reviewed_by, created_at, reviewed_at
		FROM role_upgrade_requests WHERE status = 'pending' ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách role upgrade request lỗi: %w", err)
	}
	defer rows.Close()

	var result []*roleupgrade.Request
	for rows.Next() {
		req, err := scanRoleUpgrade(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng role upgrade request lỗi: %w", err)
		}
		result = append(result, req)
	}
	return result, rows.Err()
}

func (r *RoleUpgradeRepository) Update(ctx context.Context, req *roleupgrade.Request) error {
	const q = `UPDATE role_upgrade_requests SET status = $2, reviewed_by = $3, reviewed_at = $4 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, req.ID(), string(req.Status()), req.ReviewedBy(), req.ReviewedAt())
	if err != nil {
		return fmt.Errorf("postgres: cập nhật role upgrade request lỗi: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return roleupgrade.ErrNotFound
	}
	return nil
}

func scanRoleUpgrade(row rowScanner) (*roleupgrade.Request, error) {
	var (
		id, userID uint
		reason     string
		status     string
		reviewedBy *uint
		createdAt  time.Time
		reviewedAt *time.Time
	)
	if err := row.Scan(&id, &userID, &reason, &status, &reviewedBy, &createdAt, &reviewedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, roleupgrade.ErrNotFound
		}
		return nil, err
	}
	return roleupgrade.Rehydrate(id, userID, reason, roleupgrade.Status(status), reviewedBy, createdAt, reviewedAt), nil
}
