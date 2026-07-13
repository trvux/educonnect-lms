// Package middleware chứa các chi middleware dùng chung cho handler: xác
// thực JWT và kiểm soát truy cập theo vai trò (dùng cho RBAC của US1.3 và
// các endpoint chỉ dành cho teacher/admin như tạo/duyệt khóa học).
package middleware

import (
	"context"
	"net/http"
	"strings"

	"educonnect-lms/backend/internal/domain/user"
	"educonnect-lms/backend/internal/platform/security"
)

type ctxKey string

const claimsCtxKey ctxKey = "auth_claims"

// TokenVerifier được hiện thực bởi *security.JWTIssuer.
type TokenVerifier interface {
	Verify(tokenString string) (*security.Claims, error)
}

// RequireAuth xác thực Bearer token và lưu claims đã parse vào request
// context để handler/middleware phía sau đọc.
func RequireAuth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				http.Error(w, `{"error":"thiếu bearer token"}`, http.StatusUnauthorized)
				return
			}

			claims, err := verifier.Verify(token)
			if err != nil {
				http.Error(w, `{"error":"token không hợp lệ hoặc đã hết hạn"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth giải mã Bearer token nếu có nhưng không chặn request khi
// thiếu/token không hợp lệ — dùng cho route public nhưng cần biết vai trò
// người gọi để điều chỉnh response (vd ẩn đáp án đúng của trắc nghiệm với
// học viên/khách, xem US5.1).
func OptionalAuth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if ok && token != "" {
				if claims, err := verifier.Verify(token); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), claimsCtxKey, claims))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole phải chạy sau RequireAuth. Đây là nơi hiện thực kiểm tra RBAC
// kiểu US1.3 (vd chỉ teacher/admin mới được tạo khóa học).
func RequireRole(allowed ...user.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"thiếu thông tin xác thực"}`, http.StatusUnauthorized)
				return
			}
			for _, role := range allowed {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error":"forbidden: không đủ quyền"}`, http.StatusForbidden)
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*security.Claims, bool) {
	claims, ok := ctx.Value(claimsCtxKey).(*security.Claims)
	return claims, ok
}
