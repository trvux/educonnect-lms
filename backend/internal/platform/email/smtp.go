// Package email gửi mail thật qua SMTP (US1.6 — OTP quên mật khẩu), dùng
// net/smtp của thư viện chuẩn Go, không cần thêm dependency ngoài.
package email

import (
	"context"
	"fmt"
	"net/smtp"
)

// SMTPSender implement service/auth.EmailSender bằng Gmail SMTP (hoặc bất
// kỳ SMTP server nào hỗ trợ AUTH PLAIN qua STARTTLS ở cổng 587).
type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPSender(host, port, username, password, from string) *SMTPSender {
	return &SMTPSender{host: host, port: port, username: username, password: password, from: from}
}

func (s *SMTPSender) Send(_ context.Context, to, subject, body string) error {
	addr := s.host + ":" + s.port
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s\r\n",
		s.from, to, subject, body,
	)

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("email: gửi mail thất bại: %w", err)
	}
	return nil
}
