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
	ErrInvalidEmail      = errors.New("user: email không hợp lệ")
	ErrInvalidRole       = errors.New("user: vai trò không hợp lệ")
	ErrEmptyFullName     = errors.New("user: họ tên là bắt buộc")
	ErrEmptyPasswordHash = errors.New("user: password hash là bắt buộc")
	ErrInactive          = errors.New("user: tài khoản đã bị khoá")
	ErrNotFound          = errors.New("user: không tìm thấy")
	ErrInvalidPhone      = errors.New("user: số điện thoại không hợp lệ")
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// phonePattern chấp nhận SĐT Việt Nam dạng 0xxxxxxxxx hoặc +84xxxxxxxxx.
var phonePattern = regexp.MustCompile(`^(0\d{9}|\+84\d{9})$`)

// User là aggregate root cho quản lý tài khoản & phân quyền (Epic 1).
// Các field để private: mọi thay đổi trạng thái phải đi qua behavior method
// để invariant không bao giờ bị bypass từ tầng ngoài.
type User struct {
	id           uint
	email        string
	passwordHash string
	fullName     string
	role         Role
	active       bool
	phone        string
	studentCode  string // MSSV (học viên) hoặc mã giảng viên
	avatarPath   string
	createdAt    time.Time
	updatedAt    time.Time
}

// NewUser tạo mới một tài khoản (US1.1). Password hash được set riêng
// qua SetPasswordHash để domain không phụ thuộc trực tiếp vào thư viện hash cụ thể.
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

// Rehydrate dựng lại 1 User từ dữ liệu đã lưu trong DB. Hàm này tin tưởng
// tầng lưu trữ (dữ liệu đã hợp lệ từ lúc ghi xuống), chỉ nên gọi từ
// các repository implementation.
func Rehydrate(id uint, email, passwordHash, fullName string, role Role, active bool, phone, studentCode, avatarPath string, createdAt, updatedAt time.Time) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		fullName:     fullName,
		role:         role,
		active:       active,
		phone:        phone,
		studentCode:  studentCode,
		avatarPath:   avatarPath,
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

// UpdateProfile hiện thực US1.4 (xem/cập nhật hồ sơ cá nhân). SĐT là tuỳ
// chọn (rỗng hợp lệ), nhưng nếu điền thì phải đúng định dạng VN.
func (u *User) UpdateProfile(fullName, phone, studentCode string) error {
	if fullName == "" {
		return ErrEmptyFullName
	}
	if phone != "" && !phonePattern.MatchString(phone) {
		return ErrInvalidPhone
	}
	u.fullName = fullName
	u.phone = phone
	u.studentCode = studentCode
	u.updatedAt = time.Now().UTC()
	return nil
}

func (u *User) SetAvatarPath(path string) {
	u.avatarPath = path
	u.updatedAt = time.Now().UTC()
}

// ChangeRole dùng cho US1.7 (Admin duyệt yêu cầu nâng cấp thành Giảng viên).
func (u *User) ChangeRole(role Role) error {
	if !role.Valid() {
		return ErrInvalidRole
	}
	u.role = role
	u.updatedAt = time.Now().UTC()
	return nil
}

// Deactivate dùng cho US1.3 (Admin khoá tài khoản vi phạm).
func (u *User) Deactivate() {
	u.active = false
	u.updatedAt = time.Now().UTC()
}

func (u *User) Activate() {
	u.active = true
	u.updatedAt = time.Now().UTC()
}

// CanLogin đảm bảo invariant "tài khoản phải active mới login được", dùng bởi auth service (US1.2).
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
func (u *User) Phone() string        { return u.phone }
func (u *User) StudentCode() string  { return u.studentCode }
func (u *User) AvatarPath() string   { return u.avatarPath }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// Repository là port mà tầng service phụ thuộc vào, được implement bởi
// internal/repository/postgres (dependency inversion: domain định nghĩa
// contract, infrastructure hiện thực contract đó).
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	// FindByPhone dùng cho US1.8 (quên email đăng nhập, tra cứu qua SĐT).
	FindByPhone(ctx context.Context, phone string) (*User, error)
	Update(ctx context.Context, u *User) error
}
