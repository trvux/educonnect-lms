package auth

import (
	"context"
	"errors"

	"educonnect-lms/backend/internal/domain/user"
)

var (
	ErrEmailTaken         = errors.New("auth: email already registered")
	ErrInvalidCredentials = errors.New("auth: invalid email or password")
)

// PasswordHasher and TokenIssuer are ports: the service depends only on
// these interfaces, never on bcrypt/jwt directly. internal/platform/security
// provides the concrete implementations wired in main.go.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

type TokenIssuer interface {
	Issue(userID uint, role user.Role) (string, error)
}

type Service struct {
	users  user.Repository
	hasher PasswordHasher
	tokens TokenIssuer
}

func NewService(users user.Repository, hasher PasswordHasher, tokens TokenIssuer) *Service {
	return &Service{users: users, hasher: hasher, tokens: tokens}
}

type RegisterInput struct {
	Email    string
	Password string
	FullName string
	Role     user.Role
}

// Register implements US1.1.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*user.User, error) {
	existing, err := s.users.FindByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	u, err := user.NewUser(in.Email, in.FullName, in.Role)
	if err != nil {
		return nil, err
	}
	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}
	if err := u.SetPasswordHash(hash); err != nil {
		return nil, err
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

type LoginInput struct {
	Email    string
	Password string
}

// Login implements US1.2, returning a signed JWT on success.
func (s *Service) Login(ctx context.Context, in LoginInput) (string, error) {
	u, err := s.users.FindByEmail(ctx, in.Email)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if err := u.CanLogin(); err != nil {
		return "", err
	}
	if err := s.hasher.Compare(u.PasswordHash(), in.Password); err != nil {
		return "", ErrInvalidCredentials
	}
	return s.tokens.Issue(u.ID(), u.Role())
}
