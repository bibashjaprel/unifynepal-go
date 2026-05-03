package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID  uuid.UUID `json:"user_id"`
	Email   string    `json:"email"`
	TokenID string    `json:"token_id"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, email string, secret string, expiresInHours int) (string, string, time.Time, error) {
	tokenID := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(expiresInHours) * time.Hour)

	claims := JWTClaims{
		UserID:  userID,
		Email:   email,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))

	return signedToken, tokenID, expiresAt, err
}
