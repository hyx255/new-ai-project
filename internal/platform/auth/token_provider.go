package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims for authenticated users.
type Claims struct {
	UserID             string `json:"sub"`
	Username           string `json:"username"`
	Role               string `json:"role"`
	MustChangePassword bool   `json:"must_change_password"`
	TokenVersion       int    `json:"token_version"`
	jwt.RegisteredClaims
}

// TokenProvider handles JWT generation and validation.
type TokenProvider struct {
	secret          []byte
	tokenExpiryHours int
}

// NewTokenProvider creates a new TokenProvider.
func NewTokenProvider(secret string, tokenExpiryHours int) *TokenProvider {
	if tokenExpiryHours < 1 {
		tokenExpiryHours = 2
	}
	return &TokenProvider{
		secret:          []byte(secret),
		tokenExpiryHours: tokenExpiryHours,
	}
}

// Generate creates a new JWT for the given user.
func (p *TokenProvider) Generate(userID, username, role string, mustChangePassword bool, tokenVersion int) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:             userID,
		Username:           username,
		Role:               role,
		MustChangePassword: mustChangePassword,
		TokenVersion:       tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(p.tokenExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(p.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// Parse validates and parses a JWT string, returning the claims.
func (p *TokenProvider) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return p.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}