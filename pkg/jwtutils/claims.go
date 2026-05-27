package jwtutils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims represents JWT custom claims containing user information and standard registered claims.
type CustomClaims struct {
	UserID      string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	jwt.RegisteredClaims
}

func newCustomClaims(userID, displayName, email string, expiry time.Duration, issuer, audience string) *CustomClaims {
	return &CustomClaims{
		UserID:      userID,
		DisplayName: displayName,
		Email:       email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
		},
	}
}
