package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/passwordreset"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordResetRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{pool: pool}
}

func (r *PasswordResetRepository) Create(ctx context.Context, reset *passwordreset.Reset) error {
	const q = `
		INSERT INTO password_resets (user_id, otp_hash, expires_at, attempts, used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q,
		reset.UserID(), reset.OTPHash(), reset.ExpiresAt(), reset.Attempts(), reset.Used(), reset.CreatedAt(),
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo password reset lỗi: %w", err)
	}
	reset.SetID(id)
	return nil
}

func (r *PasswordResetRepository) FindActiveByUser(ctx context.Context, userID uint) (*passwordreset.Reset, error) {
	const q = `
		SELECT id, user_id, otp_hash, expires_at, attempts, used, created_at
		FROM password_resets WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	var (
		id, uID   uint
		otpHash   string
		expiresAt time.Time
		attempts  int
		used      bool
		createdAt time.Time
	)
	err := r.pool.QueryRow(ctx, q, userID).Scan(&id, &uID, &otpHash, &expiresAt, &attempts, &used, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, passwordreset.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu password reset lỗi: %w", err)
	}
	return passwordreset.Rehydrate(id, uID, otpHash, expiresAt, attempts, used, createdAt), nil
}

func (r *PasswordResetRepository) Update(ctx context.Context, reset *passwordreset.Reset) error {
	const q = `UPDATE password_resets SET attempts = $2, used = $3 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, reset.ID(), reset.Attempts(), reset.Used())
	if err != nil {
		return fmt.Errorf("postgres: cập nhật password reset lỗi: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return passwordreset.ErrNotFound
	}
	return nil
}
