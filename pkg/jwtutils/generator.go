package jwtutils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenGenerator defines the contract for JWT token generation operations.
type TokenGenerator interface {
	GenerateToken(userID, displayName, email string) (string, error)
}

type rsaTokenGenerator struct {
	privateKey *rsa.PrivateKey
	issuer     string
	audience   string
	expiry     time.Duration
}

// NewTokenGenerator creates a new token generator with the given RSA private key, issuer, audience, and token expiry duration.
func NewTokenGenerator(privateKeyPath, issuer, audience string, expiry time.Duration) (TokenGenerator, error) {
	privateKey, err := loadRSAPrivateKeyFromFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load RSA private key: %w", err)
	}

	return &rsaTokenGenerator{
		privateKey: privateKey,
		issuer:     issuer,
		audience:   audience,
		expiry:     expiry,
	}, nil
}

// GenerateToken generates a JWT token string with the given user information and configured claims.
func (g *rsaTokenGenerator) GenerateToken(userID, displayName, email string) (string, error) {
	claims := newCustomClaims(userID, displayName, email, g.expiry, g.issuer, g.audience)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(g.privateKey)
}

func loadRSAPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSA private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM block in key file")
	}

	var privateKey *rsa.PrivateKey
	if block.Type == "RSA PRIVATE KEY" {
		// PKCS#1 format
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#1 RSA private key: %w", err)
		}
		privateKey = key
	} else if block.Type == "PRIVATE KEY" {
		// PKCS#8 format
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
		}
		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	} else {
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	return privateKey, nil
}