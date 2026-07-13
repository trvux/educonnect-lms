package passwordreset_test

import (
	"errors"
	"testing"
	"time"

	"educonnect-lms/backend/internal/domain/passwordreset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeHasher mô phỏng bcrypt: OTP đúng khi giá trị plain khớp otpHash "hash:<otp>".
type fakeHasher struct{}

func (fakeHasher) Compare(hash, plain string) error {
	if hash != "hash:"+plain {
		return errors.New("mismatch")
	}
	return nil
}

func TestReset_Verify_Success(t *testing.T) {
	r := passwordreset.New(1, "hash:123456", 10*time.Minute)
	require.NoError(t, r.Verify("123456", fakeHasher{}))
}

func TestReset_Verify_WrongOTP(t *testing.T) {
	r := passwordreset.New(1, "hash:123456", 10*time.Minute)
	err := r.Verify("000000", fakeHasher{})
	assert.ErrorIs(t, err, passwordreset.ErrInvalidOTP)
}

func TestReset_Verify_Expired(t *testing.T) {
	r := passwordreset.New(1, "hash:123456", -1*time.Minute) // đã hết hạn ngay khi tạo
	err := r.Verify("123456", fakeHasher{})
	assert.ErrorIs(t, err, passwordreset.ErrInvalidOTP)
}

func TestReset_Verify_AlreadyUsed(t *testing.T) {
	r := passwordreset.New(1, "hash:123456", 10*time.Minute)
	require.NoError(t, r.Verify("123456", fakeHasher{}))

	err := r.Verify("123456", fakeHasher{})
	assert.ErrorIs(t, err, passwordreset.ErrInvalidOTP)
}

func TestReset_Verify_TooManyAttempts(t *testing.T) {
	r := passwordreset.New(1, "hash:123456", 10*time.Minute)
	for i := 0; i < 5; i++ {
		_ = r.Verify("wrong", fakeHasher{})
	}
	err := r.Verify("123456", fakeHasher{})
	assert.ErrorIs(t, err, passwordreset.ErrTooManyAttempts)
}
