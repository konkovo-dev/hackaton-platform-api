package centrifugo

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTHelper struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTHelper(secret string, ttl time.Duration) *JWTHelper {
	return &JWTHelper{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

type CentrifugoConnectionClaims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

func (h *JWTHelper) GenerateConnectionToken(userID string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(h.ttl)

	claims := CentrifugoConnectionClaims{
		Sub: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(h.secret)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, expiresAt.Unix(), nil
}
