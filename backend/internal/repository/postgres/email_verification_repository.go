package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/emailverification"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailVerificationRepository struct {
	pool *pgxpool.Pool
}

func NewEmailVerificationRepository(pool *pgxpool.Pool) *EmailVerificationRepository {
	return &EmailVerificationRepository{pool: pool}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, v *emailverification.Verification) error {
	const q = `
		INSERT INTO email_verifications (user_id, otp_hash, expires_at, attempts, used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q,
		v.UserID(), v.OTPHash(), v.ExpiresAt(), v.Attempts(), v.Used(), v.CreatedAt(),
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo email verification lỗi: %w", err)
	}
	v.SetID(id)
	return nil
}

func (r *EmailVerificationRepository) FindActiveByUser(ctx context.Context, userID uint) (*emailverification.Verification, error) {
	const q = `
		SELECT id, user_id, otp_hash, expires_at, attempts, used, created_at
		FROM email_verifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
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
			return nil, emailverification.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu email verification lỗi: %w", err)
	}
	return emailverification.Rehydrate(id, uID, otpHash, expiresAt, attempts, used, createdAt), nil
}

func (r *EmailVerificationRepository) Update(ctx context.Context, v *emailverification.Verification) error {
	const q = `UPDATE email_verifications SET attempts = $2, used = $3 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, v.ID(), v.Attempts(), v.Used())
	if err != nil {
		return fmt.Errorf("postgres: cập nhật email verification lỗi: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return emailverification.ErrNotFound
	}
	return nil
}
