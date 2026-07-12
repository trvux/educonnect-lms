package user

import (
	"context"
	"errors"
	"regexp"
	"time"
)

type Role string

const (
	RoleStudent Role = "student"
	RoleTeacher Role = "teacher"
	RoleAdmin   Role = "admin"
)

func (r Role) Valid() bool {
	switch r {
	case RoleStudent, RoleTeacher, RoleAdmin:
		return true
	default:
		return false
	}
}

var (
	ErrInvalidEmail      = errors.New("user: invalid email")
	ErrInvalidRole       = errors.New("user: invalid role")
	ErrEmptyFullName     = errors.New("user: full name is required")
	ErrEmptyPasswordHash = errors.New("user: password hash is required")
	ErrInactive          = errors.New("user: account is deactivated")
	ErrNotFound          = errors.New("user: not found")
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// User is the aggregate root for account & authorization (Epic 1).
// Fields are private: all state changes must go through behavior methods
// so invariants can never be bypassed by outer layers.
type User struct {
	id           uint
	email        string
	passwordHash string
	fullName     string
	role         Role
	active       bool
	createdAt    time.Time
	updatedAt    time.Time
}

// NewUser creates a brand-new account (US1.1). Password hash is set separately
// via SetPasswordHash so the domain never depends on a concrete hashing library.
func NewUser(email, fullName string, role Role) (*User, error) {
	if !emailPattern.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	if fullName == "" {
		return nil, ErrEmptyFullName
	}
	if !role.Valid() {
		return nil, ErrInvalidRole
	}
	now := time.Now().UTC()
	return &User{
		email:     email,
		fullName:  fullName,
		role:      role,
		active:    true,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Rehydrate reconstructs a User from persisted data. It trusts the storage
// layer (data was already valid when it was written) and is only meant to be
// called from repository implementations.
func Rehydrate(id uint, email, passwordHash, fullName string, role Role, active bool, createdAt, updatedAt time.Time) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		fullName:     fullName,
		role:         role,
		active:       active,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (u *User) SetPasswordHash(hash string) error {
	if hash == "" {
		return ErrEmptyPasswordHash
	}
	u.passwordHash = hash
	u.updatedAt = time.Now().UTC()
	return nil
}

// Deactivate is used by US1.3 (Admin khoá tài khoản).
func (u *User) Deactivate() {
	u.active = false
	u.updatedAt = time.Now().UTC()
}

func (u *User) Activate() {
	u.active = true
	u.updatedAt = time.Now().UTC()
}

// CanLogin enforces the account-must-be-active invariant used by the auth service (US1.2).
func (u *User) CanLogin() error {
	if !u.active {
		return ErrInactive
	}
	return nil
}

func (u *User) SetID(id uint) { u.id = id }

func (u *User) ID() uint             { return u.id }
func (u *User) Email() string        { return u.email }
func (u *User) PasswordHash() string { return u.passwordHash }
func (u *User) FullName() string     { return u.fullName }
func (u *User) Role() Role           { return u.role }
func (u *User) Active() bool         { return u.active }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// Repository is the port the service layer depends on. It is implemented by
// internal/repository/postgres (dependency inversion: domain defines the
// contract, infrastructure satisfies it).
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, u *User) error
}
