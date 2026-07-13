package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"educonnect-lms/backend/internal/domain/passwordreset"
	"educonnect-lms/backend/internal/domain/user"
)

// otpTTL là thời gian hiệu lực của mã OTP quên mật khẩu (US1.6).
const otpTTL = 10 * time.Minute

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

// FileStorage là port lưu trữ file vật lý, tái dùng cho avatar (US1.4)
// giống cơ chế internal/platform/storage đã dùng cho tài liệu (US4.1).
type FileStorage interface {
	Save(ctx context.Context, subdir, fileName string, content io.Reader) (path string, err error)
}

// EmailSender là port gửi mail (US1.6), implement bởi internal/platform/email.
type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type Service struct {
	users          user.Repository
	hasher         PasswordHasher
	tokens         TokenIssuer
	storage        FileStorage
	passwordResets passwordreset.Repository
	emailSender    EmailSender
}

func NewService(
	users user.Repository,
	hasher PasswordHasher,
	tokens TokenIssuer,
	storage FileStorage,
	passwordResets passwordreset.Repository,
	emailSender EmailSender,
) *Service {
	return &Service{
		users:          users,
		hasher:         hasher,
		tokens:         tokens,
		storage:        storage,
		passwordResets: passwordResets,
		emailSender:    emailSender,
	}
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

// GetProfile hiện thực phần "xem" của US1.4.
func (s *Service) GetProfile(ctx context.Context, userID uint) (*user.User, error) {
	return s.users.FindByID(ctx, userID)
}

// UpdateProfile hiện thực phần "cập nhật" của US1.4.
func (s *Service) UpdateProfile(ctx context.Context, userID uint, fullName, phone, studentCode string) (*user.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := u.UpdateProfile(fullName, phone, studentCode); err != nil {
		return nil, err
	}
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// ChangePassword hiện thực US1.5 (đã đăng nhập, biết mật khẩu cũ).
func (s *Service) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.hasher.Compare(u.PasswordHash(), currentPassword); err != nil {
		return ErrInvalidCredentials
	}
	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	if err := u.SetPasswordHash(hash); err != nil {
		return err
	}
	return s.users.Update(ctx, u)
}

// UploadAvatar hiện thực phần upload ảnh đại diện của US1.4.
func (s *Service) UploadAvatar(ctx context.Context, userID uint, fileName string, content io.Reader) (*user.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	subdir := "avatars/" + strconv.FormatUint(uint64(userID), 10)
	path, err := s.storage.Save(ctx, subdir, fileName, content)
	if err != nil {
		return nil, err
	}
	u.SetAvatarPath(path)
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// ForgotUsername hiện thực US1.8: tra cứu email đăng nhập qua SĐT, trả về
// email đã che bớt (không lộ toàn bộ) thay vì email thật.
func (s *Service) ForgotUsername(ctx context.Context, phone string) (string, error) {
	u, err := s.users.FindByPhone(ctx, phone)
	if err != nil {
		return "", err
	}
	return maskEmail(u.Email()), nil
}

// ForgotPassword hiện thực US1.6: sinh OTP, lưu (đã hash) và gửi qua email.
// Luôn trả nil dù email không tồn tại — không tiết lộ email nào đã đăng ký
// (chống dò tài khoản), trừ khi có lỗi hệ thống thật sự.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil
		}
		return err
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}
	otpHash, err := s.hasher.Hash(otp)
	if err != nil {
		return err
	}

	reset := passwordreset.New(u.ID(), otpHash, otpTTL)
	if err := s.passwordResets.Create(ctx, reset); err != nil {
		return err
	}

	body := fmt.Sprintf(
		"Mã OTP đặt lại mật khẩu EduConnect LMS của bạn là: %s\nMã có hiệu lực trong 10 phút. Nếu bạn không yêu cầu, hãy bỏ qua email này.",
		otp,
	)
	return s.emailSender.Send(ctx, u.Email(), "EduConnect LMS - Mã OTP đặt lại mật khẩu", body)
}

// ResetPassword hiện thực US1.6: xác thực OTP rồi đặt mật khẩu mới.
func (s *Service) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return passwordreset.ErrInvalidOTP // không tiết lộ email có tồn tại hay không
	}

	reset, err := s.passwordResets.FindActiveByUser(ctx, u.ID())
	if err != nil {
		return passwordreset.ErrInvalidOTP
	}

	verifyErr := reset.Verify(otp, s.hasher)
	if updateErr := s.passwordResets.Update(ctx, reset); updateErr != nil {
		return updateErr
	}
	if verifyErr != nil {
		return verifyErr
	}

	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	if err := u.SetPasswordHash(hash); err != nil {
		return err
	}
	return s.users.Update(ctx, u)
}

// generateOTP sinh mã 6 chữ số bằng crypto/rand (an toàn hơn math/rand cho
// mục đích bảo mật, dù chỉ là OTP ngắn hạn).
func generateOTP() (string, error) {
	const digits = "0123456789"
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	otp := make([]byte, 6)
	for i, v := range b {
		otp[i] = digits[int(v)%len(digits)]
	}
	return string(otp), nil
}

// maskEmail che phần local-part của email, giữ lại tối đa 2 ký tự đầu, vd
// "huy@vlu.edu.vn" -> "hu***@vlu.edu.vn".
func maskEmail(email string) string {
	at := strings.IndexByte(email, '@')
	if at <= 0 {
		return email
	}
	local, domain := email[:at], email[at:]
	visible := local
	if len(visible) > 2 {
		visible = visible[:2]
	}
	return visible + "***" + domain
}
