package emailverification_test

import (
	"errors"
	"testing"
	"time"

	"educonnect-lms/backend/internal/domain/emailverification"

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

func TestVerification_Verify_Success(t *testing.T) {
	v := emailverification.New(1, "hash:123456", 10*time.Minute)
	require.NoError(t, v.Verify("123456", fakeHasher{}))
}

func TestVerification_Verify_WrongOTP(t *testing.T) {
	v := emailverification.New(1, "hash:123456", 10*time.Minute)
	err := v.Verify("000000", fakeHasher{})
	assert.ErrorIs(t, err, emailverification.ErrInvalidOTP)
}

func TestVerification_Verify_Expired(t *testing.T) {
	v := emailverification.New(1, "hash:123456", -1*time.Minute) // đã hết hạn ngay khi tạo
	err := v.Verify("123456", fakeHasher{})
	assert.ErrorIs(t, err, emailverification.ErrInvalidOTP)
}

func TestVerification_Verify_AlreadyUsed(t *testing.T) {
	v := emailverification.New(1, "hash:123456", 10*time.Minute)
	require.NoError(t, v.Verify("123456", fakeHasher{}))

	err := v.Verify("123456", fakeHasher{})
	assert.ErrorIs(t, err, emailverification.ErrInvalidOTP)
}

func TestVerification_Verify_TooManyAttempts(t *testing.T) {
	v := emailverification.New(1, "hash:123456", 10*time.Minute)
	for i := 0; i < 5; i++ {
		_ = v.Verify("wrong", fakeHasher{})
	}
	err := v.Verify("123456", fakeHasher{})
	assert.ErrorIs(t, err, emailverification.ErrTooManyAttempts)
}
