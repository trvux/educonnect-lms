// Package middleware holds chi middleware shared across handlers: JWT
// verification and role-based access control (used by US1.3's RBAC and by
// teacher/admin-only endpoints such as course creation and approval).
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

// TokenVerifier is satisfied by *security.JWTIssuer.
type TokenVerifier interface {
	Verify(tokenString string) (*security.Claims, error)
}

// RequireAuth verifies the Bearer token and stores the parsed claims on the
// request context for downstream handlers/middleware to read.
func RequireAuth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}

			claims, err := verifier.Verify(token)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole must run after RequireAuth. It implements the RBAC checks for
// US1.3-style role restrictions (e.g. only teacher/admin can create courses).
func RequireRole(allowed ...user.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"missing auth context"}`, http.StatusUnauthorized)
				return
			}
			for _, role := range allowed {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error":"forbidden: insufficient role"}`, http.StatusForbidden)
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*security.Claims, bool) {
	claims, ok := ctx.Value(claimsCtxKey).(*security.Claims)
	return claims, ok
}
