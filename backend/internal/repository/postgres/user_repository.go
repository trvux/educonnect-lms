package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/user"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	const q = `
		INSERT INTO users (email, password_hash, full_name, role, active, email_verified, phone, student_code, avatar_path, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q,
		u.Email(), u.PasswordHash(), u.FullName(), u.Role(), u.Active(), u.EmailVerified(), u.Phone(), u.StudentCode(), u.AvatarPath(), u.CreatedAt(), u.UpdatedAt(),
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: tạo user lỗi: %w", err)
	}
	u.SetID(id)
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, role, active, email_verified, phone, student_code, avatar_path, created_at, updated_at
		FROM users WHERE email = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, email))
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*user.User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, role, active, email_verified, phone, student_code, avatar_path, created_at, updated_at
		FROM users WHERE id = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, id))
}

// FindByPhone dùng cho US1.8 (quên email đăng nhập, tra cứu qua SĐT).
func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*user.User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, role, active, email_verified, phone, student_code, avatar_path, created_at, updated_at
		FROM users WHERE phone = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, phone))
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	const q = `
		UPDATE users SET email = $1, password_hash = $2, full_name = $3, role = $4, active = $5,
			email_verified = $6, phone = $7, student_code = $8, avatar_path = $9, updated_at = $10
		WHERE id = $11`
	tag, err := r.pool.Exec(ctx, q,
		u.Email(), u.PasswordHash(), u.FullName(), u.Role(), u.Active(), u.EmailVerified(), u.Phone(), u.StudentCode(), u.AvatarPath(), u.UpdatedAt(), u.ID(),
	)
	if err != nil {
		return fmt.Errorf("postgres: cập nhật user lỗi: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *UserRepository) scanOne(row pgx.Row) (*user.User, error) {
	var (
		id                             uint
		email, hash, fullName          string
		role                           user.Role
		active, emailVerified          bool
		phone, studentCode, avatarPath string
		createdAt, updatedAt           time.Time
	)
	err := row.Scan(&id, &email, &hash, &fullName, &role, &active, &emailVerified, &phone, &studentCode, &avatarPath, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu user lỗi: %w", err)
	}
	return user.Rehydrate(id, email, hash, fullName, role, active, emailVerified, phone, studentCode, avatarPath, createdAt, updatedAt), nil
}
