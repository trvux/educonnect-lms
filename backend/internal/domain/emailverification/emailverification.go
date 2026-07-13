// Package emailverification là domain của US1.9: xác thực quyền sở hữu
// email lúc đăng ký qua mã OTP gửi email, trước khi tài khoản được active
// để đăng nhập.
package emailverification

import (
	"context"
	"errors"
	"time"
)

const maxAttempts = 5

var (
	ErrInvalidOTP      = errors.New("emailverification: mã OTP không đúng hoặc đã hết hạn")
	ErrTooManyAttempts = errors.New("emailverification: đã vượt quá số lần thử OTP cho phép")
	ErrNotFound        = errors.New("emailverification: không tìm thấy")
)

// OTPHasher so khớp OTP người dùng nhập với hash đã lưu — cùng interface
// shape với auth.PasswordHasher nên có thể tái dùng bcrypt hasher sẵn có.
type OTPHasher interface {
	Compare(hash, plain string) error
}

// Verification là 1 yêu cầu xác thực email qua OTP, gắn với 1 user cụ thể.
type Verification struct {
	id        uint
	userID    uint
	otpHash   string
	expiresAt time.Time
	attempts  int
	used      bool
	createdAt time.Time
}

func New(userID uint, otpHash string, ttl time.Duration) *Verification {
	now := time.Now().UTC()
	return &Verification{userID: userID, otpHash: otpHash, expiresAt: now.Add(ttl), createdAt: now}
}

func Rehydrate(id, userID uint, otpHash string, expiresAt time.Time, attempts int, used bool, createdAt time.Time) *Verification {
	return &Verification{id: id, userID: userID, otpHash: otpHash, expiresAt: expiresAt, attempts: attempts, used: used, createdAt: createdAt}
}

// Verify kiểm tra OTP: hết hạn/đã dùng/vượt số lần thử đều bị chặn. Số lần
// thử luôn tăng dù đúng hay sai (kể cả không khớp OTP) để chống brute-force.
func (v *Verification) Verify(otp string, hasher OTPHasher) error {
	if v.used || time.Now().UTC().After(v.expiresAt) {
		return ErrInvalidOTP
	}
	if v.attempts >= maxAttempts {
		return ErrTooManyAttempts
	}
	v.attempts++
	if err := hasher.Compare(v.otpHash, otp); err != nil {
		return ErrInvalidOTP
	}
	v.used = true
	return nil
}

func (v *Verification) SetID(id uint) { v.id = id }

func (v *Verification) ID() uint             { return v.id }
func (v *Verification) UserID() uint         { return v.userID }
func (v *Verification) OTPHash() string      { return v.otpHash }
func (v *Verification) Attempts() int        { return v.attempts }
func (v *Verification) Used() bool           { return v.used }
func (v *Verification) ExpiresAt() time.Time { return v.expiresAt }
func (v *Verification) CreatedAt() time.Time { return v.createdAt }

type Repository interface {
	Create(ctx context.Context, v *Verification) error
	// FindActiveByUser trả về yêu cầu xác thực gần nhất của user đó.
	FindActiveByUser(ctx context.Context, userID uint) (*Verification, error)
	Update(ctx context.Context, v *Verification) error
}
