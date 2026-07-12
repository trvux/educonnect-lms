package auth

import (
	"context"
	"errors"

	"educonnect-lms/backend/internal/domain/user"
)

var (
	ErrEmailTaken         = errors.New("auth: email đã được đăng ký")
	ErrInvalidCredentials = errors.New("auth: email hoặc mật khẩu không đúng")
)

// PasswordHasher và TokenIssuer là các port: service chỉ phụ thuộc vào
// interface, không phụ thuộc trực tiếp bcrypt/jwt. internal/platform/security
// cung cấp implementation cụ thể, được wire ở main.go.
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

// Register hiện thực US1.1.
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

// Login hiện thực US1.2, trả về JWT đã ký nếu đăng nhập thành công.
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
