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
		INSERT INTO users (email, password_hash, full_name, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	var id uint
	err := r.pool.QueryRow(ctx, q,
		u.Email(), u.PasswordHash(), u.FullName(), u.Role(), u.Active(), u.CreatedAt(), u.UpdatedAt(),
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: create user: %w", err)
	}
	u.SetID(id)
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, role, active, created_at, updated_at
		FROM users WHERE email = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, email))
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*user.User, error) {
	const q = `
		SELECT id, email, password_hash, full_name, role, active, created_at, updated_at
		FROM users WHERE id = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, id))
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	const q = `
		UPDATE users SET email = $1, password_hash = $2, full_name = $3, role = $4, active = $5, updated_at = $6
		WHERE id = $7`
	tag, err := r.pool.Exec(ctx, q, u.Email(), u.PasswordHash(), u.FullName(), u.Role(), u.Active(), u.UpdatedAt(), u.ID())
	if err != nil {
		return fmt.Errorf("postgres: update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *UserRepository) scanOne(row pgx.Row) (*user.User, error) {
	var (
		id                    uint
		email, hash, fullName string
		role                  user.Role
		active                bool
		createdAt, updatedAt  time.Time
	)
	err := row.Scan(&id, &email, &hash, &fullName, &role, &active, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: scan user: %w", err)
	}
	return user.Rehydrate(id, email, hash, fullName, role, active, createdAt, updatedAt), nil
}
