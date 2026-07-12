package security

import (
	"errors"
	"time"

	"educonnect-lms/backend/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("security: token không hợp lệ hoặc đã hết hạn")

// Claims chứa trong JWT, được expose ra tầng handler/middleware.
type Claims struct {
	UserID uint      `json:"uid"`
	Role   user.Role `json:"role"`
	jwt.RegisteredClaims
}

// JWTIssuer hiện thực service/auth.TokenIssuer, đồng thời được HTTP auth
// middleware dùng để verify token của request.
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
