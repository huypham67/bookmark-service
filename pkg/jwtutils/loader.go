package jwtutils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadRSAPrivateKeyFromFile loads an RSA private key from a PEM-encoded file.
func LoadRSAPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSA private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM block in key file")
	}

	var privateKey *rsa.PrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#1 RSA private key: %w", err)
		}
		privateKey = key
	case "PRIVATE KEY":
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
		}
		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	return privateKey, nil
}

// LoadRSAPublicKeyFromFile loads an RSA public key from a PEM-encoded file.
func LoadRSAPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSA public key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM block in key file")
	}

	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM block type for public key: %s (expected PUBLIC KEY)", block.Type)
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPublicKey, nil
}
