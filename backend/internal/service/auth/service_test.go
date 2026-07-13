package auth_test

import (
	"context"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"educonnect-lms/backend/internal/domain/emailverification"
	"educonnect-lms/backend/internal/domain/passwordreset"
	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/service/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUserRepo is an in-memory stand-in for user.Repository so the service
// layer can be unit tested without a real Postgres instance.
type fakeUserRepo struct {
	byEmail map[string]*user.User
	nextID  uint
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]*user.User{}}
}

func (r *fakeUserRepo) Create(_ context.Context, u *user.User) error {
	r.nextID++
	u.SetID(r.nextID)
	r.byEmail[u.Email()] = u
	return nil
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if u, ok := r.byEmail[email]; ok {
		return u, nil
	}
	return nil, user.ErrNotFound
}

func (r *fakeUserRepo) FindByPhone(_ context.Context, phone string) (*user.User, error) {
	for _, u := range r.byEmail {
		if u.Phone() == phone {
			return u, nil
		}
	}
	return nil, user.ErrNotFound
}

func (r *fakeUserRepo) FindByID(_ context.Context, id uint) (*user.User, error) {
	for _, u := range r.byEmail {
		if u.ID() == id {
			return u, nil
		}
	}
	return nil, user.ErrNotFound
}

func (r *fakeUserRepo) Update(_ context.Context, u *user.User) error {
	r.byEmail[u.Email()] = u
	return nil
}

// fakeHasher avoids pulling real bcrypt into the unit test (kept fast + pure).
type fakeHasher struct{}

func (fakeHasher) Hash(plain string) (string, error) { return "hashed:" + plain, nil }
func (fakeHasher) Compare(hash, plain string) error {
	if hash != "hashed:"+plain {
		return errors.New("mismatch")
	}
	return nil
}

type fakeTokenIssuer struct{}

func (fakeTokenIssuer) Issue(userID uint, role user.Role) (string, error) {
	return "token-for-user", nil
}

type fakeStorage struct{ savedPath string }

func (s *fakeStorage) Save(_ context.Context, subdir, fileName string, content io.Reader) (string, error) {
	_, _ = io.ReadAll(content)
	s.savedPath = subdir + "/" + fileName
	return s.savedPath, nil
}

// fakePasswordResetRepo is an in-memory stand-in for passwordreset.Repository.
type fakePasswordResetRepo struct {
	byUser map[uint]*passwordreset.Reset
	nextID uint
}

func newFakePasswordResetRepo() *fakePasswordResetRepo {
	return &fakePasswordResetRepo{byUser: map[uint]*passwordreset.Reset{}}
}

func (r *fakePasswordResetRepo) Create(_ context.Context, reset *passwordreset.Reset) error {
	r.nextID++
	reset.SetID(r.nextID)
	r.byUser[reset.UserID()] = reset
	return nil
}

func (r *fakePasswordResetRepo) FindActiveByUser(_ context.Context, userID uint) (*passwordreset.Reset, error) {
	if reset, ok := r.byUser[userID]; ok {
		return reset, nil
	}
	return nil, passwordreset.ErrNotFound
}

func (r *fakePasswordResetRepo) Update(_ context.Context, reset *passwordreset.Reset) error {
	r.byUser[reset.UserID()] = reset
	return nil
}

// fakeEmailVerificationRepo is an in-memory stand-in for emailverification.Repository.
type fakeEmailVerificationRepo struct {
	byUser map[uint]*emailverification.Verification
	nextID uint
}

func newFakeEmailVerificationRepo() *fakeEmailVerificationRepo {
	return &fakeEmailVerificationRepo{byUser: map[uint]*emailverification.Verification{}}
}

func (r *fakeEmailVerificationRepo) Create(_ context.Context, v *emailverification.Verification) error {
	r.nextID++
	v.SetID(r.nextID)
	r.byUser[v.UserID()] = v
	return nil
}

func (r *fakeEmailVerificationRepo) FindActiveByUser(_ context.Context, userID uint) (*emailverification.Verification, error) {
	if v, ok := r.byUser[userID]; ok {
		return v, nil
	}
	return nil, emailverification.ErrNotFound
}

func (r *fakeEmailVerificationRepo) Update(_ context.Context, v *emailverification.Verification) error {
	r.byUser[v.UserID()] = v
	return nil
}

// fakeEmailSender captures the last email "sent" so tests can assert on it
// without touching a real SMTP server.
type fakeEmailSender struct {
	lastTo      string
	lastSubject string
	lastBody    string
}

func (s *fakeEmailSender) Send(_ context.Context, to, subject, body string) error {
	s.lastTo, s.lastSubject, s.lastBody = to, subject, body
	return nil
}

func newService() *auth.Service {
	return auth.NewService(
		newFakeUserRepo(), fakeHasher{}, fakeTokenIssuer{}, &fakeStorage{},
		newFakePasswordResetRepo(), newFakeEmailVerificationRepo(), &fakeEmailSender{},
	)
}

// newServiceForOTP trả thêm repo/mailer giả để test kiểm tra được nội dung
// OTP thật sự đã gửi (đọc từ email giả, không đoán mò).
func newServiceForOTP() (*auth.Service, *fakeEmailSender) {
	mailer := &fakeEmailSender{}
	svc := auth.NewService(
		newFakeUserRepo(), fakeHasher{}, fakeTokenIssuer{}, &fakeStorage{},
		newFakePasswordResetRepo(), newFakeEmailVerificationRepo(), mailer,
	)
	return svc, mailer
}

func TestService_Register(t *testing.T) {
	s := newService()
	ctx := context.Background()

	u, err := s.Register(ctx, auth.RegisterInput{
		Email:                 "huy@vlu.edu.vn",
		Password:              "secret123",
		FullName:              "Huynh Bao Huy",
		Role:                  user.RoleStudent,
		SkipEmailVerification: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "huy@vlu.edu.vn", u.Email())
	assert.NotEmpty(t, u.PasswordHash())

	// Registering the same email twice must fail (US1.1 uniqueness rule).
	_, err = s.Register(ctx, auth.RegisterInput{
		Email:    "huy@vlu.edu.vn",
		Password: "other",
		FullName: "Duplicate",
		Role:     user.RoleStudent,
	})
	assert.ErrorIs(t, err, auth.ErrEmailTaken)
}

func TestService_GetProfile(t *testing.T) {
	s := newService()
	ctx := context.Background()

	registered, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
		SkipEmailVerification: true,
	})
	require.NoError(t, err)

	got, err := s.GetProfile(ctx, registered.ID())
	require.NoError(t, err)
	assert.Equal(t, "huy@vlu.edu.vn", got.Email())
}

func TestService_UpdateProfile(t *testing.T) {
	s := newService()
	ctx := context.Background()

	registered, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
		SkipEmailVerification: true,
	})
	require.NoError(t, err)

	updated, err := s.UpdateProfile(ctx, registered.ID(), "Huynh Bao Huy", "0987654321", "2074802010001")
	require.NoError(t, err)
	assert.Equal(t, "Huynh Bao Huy", updated.FullName())
	assert.Equal(t, "0987654321", updated.Phone())

	_, err = s.UpdateProfile(ctx, registered.ID(), "Huy", "sai-dinh-dang", "")
	assert.ErrorIs(t, err, user.ErrInvalidPhone)
}

func TestService_ChangePassword(t *testing.T) {
	s := newService()
	ctx := context.Background()

	registered, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
		SkipEmailVerification: true,
	})
	require.NoError(t, err)

	err = s.ChangePassword(ctx, registered.ID(), "wrong-current", "new-secret")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)

	err = s.ChangePassword(ctx, registered.ID(), "secret123", "new-secret")
	require.NoError(t, err)

	_, err = s.Login(ctx, auth.LoginInput{Email: "huy@vlu.edu.vn", Password: "new-secret"})
	require.NoError(t, err)
}

func TestService_UploadAvatar(t *testing.T) {
	s := newService()
	ctx := context.Background()

	registered, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
		SkipEmailVerification: true,
	})
	require.NoError(t, err)

	updated, err := s.UploadAvatar(ctx, registered.ID(), "photo.jpg", strings.NewReader("fake image bytes"))
	require.NoError(t, err)
	assert.Equal(t, "avatars/"+strconv.FormatUint(uint64(registered.ID()), 10)+"/photo.jpg", updated.AvatarPath())
}

func TestService_ForgotUsername(t *testing.T) {
	s := newService()
	ctx := context.Background()

	registered, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
		SkipEmailVerification: true,
	})
	require.NoError(t, err)
	_, err = s.UpdateProfile(ctx, registered.ID(), "Huy", "0987654321", "")
	require.NoError(t, err)

	masked, err := s.ForgotUsername(ctx, "0987654321")
	require.NoError(t, err)
	assert.Equal(t, "hu***@vlu.edu.vn", masked)

	_, err = s.ForgotUsername(ctx, "0900000000")
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestService_ForgotAndResetPassword(t *testing.T) {
	s, mailer := newServiceForOTP()
	ctx := context.Background()

	registered, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
		SkipEmailVerification: true,
	})
	require.NoError(t, err)

	require.NoError(t, s.ForgotPassword(ctx, "huy@vlu.edu.vn"))
	require.NotEmpty(t, mailer.lastBody)
	assert.Equal(t, "huy@vlu.edu.vn", mailer.lastTo)

	// Trích OTP 6 số từ nội dung mail giả để test round-trip thật, không
	// đoán mò giá trị (OTP sinh ngẫu nhiên bằng crypto/rand).
	otp := extractOTP(t, mailer.lastBody)

	err = s.ResetPassword(ctx, "huy@vlu.edu.vn", "000000", "new-secret")
	if otp == "000000" {
		t.Fatal("OTP ngẫu nhiên trùng giá trị test dùng để thử sai — chạy lại test")
	}
	assert.ErrorIs(t, err, passwordreset.ErrInvalidOTP)

	require.NoError(t, s.ResetPassword(ctx, "huy@vlu.edu.vn", otp, "new-secret"))

	_, err = s.Login(ctx, auth.LoginInput{Email: "huy@vlu.edu.vn", Password: "new-secret"})
	require.NoError(t, err)
	_ = registered
}

func TestService_ForgotPassword_UnknownEmail_DoesNotLeak(t *testing.T) {
	s, mailer := newServiceForOTP()

	err := s.ForgotPassword(context.Background(), "nobody@vlu.edu.vn")
	require.NoError(t, err, "không được lộ thông tin email có tồn tại hay không")
	assert.Empty(t, mailer.lastTo, "không được gửi mail cho email không tồn tại")
}

func TestService_Register_RequiresEmailVerification(t *testing.T) {
	s, mailer := newServiceForOTP()
	ctx := context.Background()

	registered, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
	})
	require.NoError(t, err)
	assert.False(t, registered.EmailVerified(), "tài khoản mới đăng ký phải chưa xác thực email")
	require.NotEmpty(t, mailer.lastBody, "phải gửi OTP xác thực ngay khi đăng ký")
	assert.Equal(t, "huy@vlu.edu.vn", mailer.lastTo)

	_, err = s.Login(ctx, auth.LoginInput{Email: "huy@vlu.edu.vn", Password: "secret123"})
	assert.ErrorIs(t, err, user.ErrEmailNotVerified, "chưa xác thực email thì chưa được đăng nhập")

	otp := extractOTP(t, mailer.lastBody)
	require.NoError(t, s.VerifyEmail(ctx, "huy@vlu.edu.vn", otp))

	token, err := s.Login(ctx, auth.LoginInput{Email: "huy@vlu.edu.vn", Password: "secret123"})
	require.NoError(t, err)
	assert.Equal(t, "token-for-user", token)

	// Verify lại lần 2 (vd double-submit) phải idempotent, không lỗi.
	require.NoError(t, s.VerifyEmail(ctx, "huy@vlu.edu.vn", otp))
}

func TestService_VerifyEmail_WrongOTP(t *testing.T) {
	s, mailer := newServiceForOTP()
	ctx := context.Background()

	_, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
	})
	require.NoError(t, err)

	otp := extractOTP(t, mailer.lastBody)
	wrong := "000000"
	if wrong == otp {
		wrong = "111111"
	}
	err = s.VerifyEmail(ctx, "huy@vlu.edu.vn", wrong)
	assert.ErrorIs(t, err, emailverification.ErrInvalidOTP)

	_, err = s.Login(ctx, auth.LoginInput{Email: "huy@vlu.edu.vn", Password: "secret123"})
	assert.ErrorIs(t, err, user.ErrEmailNotVerified)
}

func TestService_ResendVerification(t *testing.T) {
	s, mailer := newServiceForOTP()
	ctx := context.Background()

	_, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
	})
	require.NoError(t, err)
	firstOTP := extractOTP(t, mailer.lastBody)

	require.NoError(t, s.ResendVerification(ctx, "huy@vlu.edu.vn"))
	secondOTP := extractOTP(t, mailer.lastBody)

	require.NoError(t, s.VerifyEmail(ctx, "huy@vlu.edu.vn", secondOTP))
	_ = firstOTP

	_, err = s.Login(ctx, auth.LoginInput{Email: "huy@vlu.edu.vn", Password: "secret123"})
	require.NoError(t, err)
}

func TestService_ResendVerification_UnknownEmail_DoesNotLeak(t *testing.T) {
	s, mailer := newServiceForOTP()

	err := s.ResendVerification(context.Background(), "nobody@vlu.edu.vn")
	require.NoError(t, err, "không được lộ thông tin email có tồn tại hay không")
	assert.Empty(t, mailer.lastTo, "không được gửi mail cho email không tồn tại")
}

func extractOTP(t *testing.T, body string) string {
	t.Helper()
	re := regexp.MustCompile(`\d{6}`)
	match := re.FindString(body)
	require.NotEmpty(t, match, "không tìm thấy OTP 6 số trong nội dung mail: %s", body)
	return match
}

func TestService_Login(t *testing.T) {
	s := newService()
	ctx := context.Background()

	_, err := s.Register(ctx, auth.RegisterInput{
		Email: "huy@vlu.edu.vn", Password: "secret123", FullName: "Huy", Role: user.RoleStudent,
		SkipEmailVerification: true,
	})
	require.NoError(t, err)

	t.Run("correct credentials", func(t *testing.T) {
		token, err := s.Login(ctx, auth.LoginInput{Email: "huy@vlu.edu.vn", Password: "secret123"})
		require.NoError(t, err)
		assert.Equal(t, "token-for-user", token)
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := s.Login(ctx, auth.LoginInput{Email: "huy@vlu.edu.vn", Password: "wrong"})
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	})

	t.Run("unknown email", func(t *testing.T) {
		_, err := s.Login(ctx, auth.LoginInput{Email: "nobody@vlu.edu.vn", Password: "x"})
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	})
}
