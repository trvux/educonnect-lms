// Package passwordreset là domain của US1.6: Học viên/Giảng viên quên mật
// khẩu, đặt lại qua mã OTP gửi email.
package passwordreset

import (
	"context"
	"errors"
	"time"
)

const maxAttempts = 5

var (
	ErrInvalidOTP      = errors.New("passwordreset: mã OTP không đúng hoặc đã hết hạn")
	ErrTooManyAttempts = errors.New("passwordreset: đã vượt quá số lần thử OTP cho phép")
	ErrNotFound        = errors.New("passwordreset: không tìm thấy")
)

// OTPHasher so khớp OTP người dùng nhập với hash đã lưu — cùng interface
// shape với auth.PasswordHasher nên có thể tái dùng bcrypt hasher sẵn có.
type OTPHasher interface {
	Compare(hash, plain string) error
}

// Reset là 1 yêu cầu đặt lại mật khẩu qua OTP, gắn với 1 user cụ thể.
type Reset struct {
	id        uint
	userID    uint
	otpHash   string
	expiresAt time.Time
	attempts  int
	used      bool
	createdAt time.Time
}

func New(userID uint, otpHash string, ttl time.Duration) *Reset {
	now := time.Now().UTC()
	return &Reset{userID: userID, otpHash: otpHash, expiresAt: now.Add(ttl), createdAt: now}
}

func Rehydrate(id, userID uint, otpHash string, expiresAt time.Time, attempts int, used bool, createdAt time.Time) *Reset {
	return &Reset{id: id, userID: userID, otpHash: otpHash, expiresAt: expiresAt, attempts: attempts, used: used, createdAt: createdAt}
}

// Verify kiểm tra OTP: hết hạn/đã dùng/vượt số lần thử đều bị chặn. Số lần
// thử luôn tăng dù đúng hay sai (kể cả không khớp OTP) để chống brute-force.
func (r *Reset) Verify(otp string, hasher OTPHasher) error {
	if r.used || time.Now().UTC().After(r.expiresAt) {
		return ErrInvalidOTP
	}
	if r.attempts >= maxAttempts {
		return ErrTooManyAttempts
	}
	r.attempts++
	if err := hasher.Compare(r.otpHash, otp); err != nil {
		return ErrInvalidOTP
	}
	r.used = true
	return nil
}

func (r *Reset) SetID(id uint) { r.id = id }

func (r *Reset) ID() uint             { return r.id }
func (r *Reset) UserID() uint         { return r.userID }
func (r *Reset) OTPHash() string      { return r.otpHash }
func (r *Reset) Attempts() int        { return r.attempts }
func (r *Reset) Used() bool           { return r.used }
func (r *Reset) ExpiresAt() time.Time { return r.expiresAt }
func (r *Reset) CreatedAt() time.Time { return r.createdAt }

type Repository interface {
	Create(ctx context.Context, r *Reset) error
	// FindActiveByUser trả về yêu cầu đặt lại mật khẩu gần nhất của user đó.
	FindActiveByUser(ctx context.Context, userID uint) (*Reset, error)
	Update(ctx context.Context, r *Reset) error
}
