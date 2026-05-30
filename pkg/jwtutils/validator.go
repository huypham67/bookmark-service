package jwtutils

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// TokenValidator defines the contract for JWT token validation operations.
// mockery --name=TokenValidator --dir=pkg/jwtutils --output=pkg/jwtutils/mocks --filename=validator.go
type TokenValidator interface {
	ValidateToken(tokenString string) (*CustomClaims, error)
}

type rsaTokenValidator struct {
	publicKey *rsa.PublicKey
	issuer    string
	audience  string
}

// NewTokenValidator creates a new token validator with the given RSA public key, issuer, and audience.
func NewTokenValidator(publicKey *rsa.PublicKey, issuer, audience string) (TokenValidator, error) {
	return &rsaTokenValidator{
		publicKey: publicKey,
		issuer:    issuer,
		audience:  audience,
	}, nil
}

// ValidateToken validates the given JWT token string and returns the custom claims if the token is valid.
func (v *rsaTokenValidator) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		if claims.Issuer != v.issuer {
			return nil, fmt.Errorf("invalid token issuer")
		}
		// Check if audience is in the token's audience claims
		audienceFound := false
		for _, aud := range claims.Audience {
			if aud == v.audience {
				audienceFound = true
				break
			}
		}
		if !audienceFound {
			return nil, fmt.Errorf("invalid token audience")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
