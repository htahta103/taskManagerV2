package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

// MintAccessToken issues a signed JWT with sub=userID.
func MintAccessToken(userID uuid.UUID, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("empty jwt secret")
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// ParseAccessToken validates HS256 JWT and returns the user id from sub.
func ParseAccessToken(tokenString, secret string) (uuid.UUID, error) {
	if secret == "" {
		return uuid.Nil, ErrInvalidToken
	}
	var claims jwt.RegisteredClaims
	_, err := jwt.NewParser().ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	if claims.Subject == "" {
		return uuid.Nil, ErrInvalidToken
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return id, nil
}
