-- US1.9: tài khoản mới cần xác thực OTP trước khi active để login. Users đã
-- có từ trước migration này được grandfathered là đã xác thực (default true)
-- để không bị khoá đăng nhập đột ngột.
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT true;
