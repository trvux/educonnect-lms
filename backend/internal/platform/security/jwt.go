package security

import (
	"errors"
	"time"

	"educonnect-lms/backend/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("security: invalid or expired token")

// Claims carried inside the JWT, exposed to the handler/middleware layer.
type Claims struct {
	UserID uint      `json:"uid"`
	Role   user.Role `json:"role"`
	jwt.RegisteredClaims
}

// JWTIssuer implements service/auth.TokenIssuer and is also used by the
// HTTP auth middleware to verify incoming tokens.
type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTIssuer(secret string, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret), ttl: ttl}
}

func (j *JWTIssuer) Issue(userID uint, role user.Role) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTIssuer) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
