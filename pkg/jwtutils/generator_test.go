package jwtutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempPrivateKeyFile(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	tmpFile, err := os.CreateTemp("", "private-key-*.pem")
	require.NoError(t, err)

	err = os.WriteFile(tmpFile.Name(), pemData, 0600)
	require.NoError(t, err)

	return tmpFile.Name()
}

func TestRSATokenGenerator_GenerateToken(t *testing.T) {
	t.Parallel()

	type fields struct {
		issuer   string
		audience string
		expiry   time.Duration
	}

	type args struct {
		userID      string
		displayName string
		email       string
	}

	testCases := []struct {
		name   string
		fields fields
		args   args
		verify func(*testing.T, string, error, args, fields)
	}{
		{
			name: "should generate token with correct claims",
			fields: fields{
				issuer:   "bookmark-service",
				audience: "bookmark-client",
				expiry:   time.Hour,
			},
			args: args{
				userID:      "user-123",
				displayName: "John Doe",
				email:       "john@example.com",
			},
			verify: func(t *testing.T, tokenString string, err error, a args, f fields) {
				require.NoError(t, err)
				require.NotEmpty(t, tokenString)

				token, _, err := jwt.NewParser().ParseUnverified(
					tokenString,
					&CustomClaims{},
				)
				require.NoError(t, err)

				claims, ok := token.Claims.(*CustomClaims)
				require.True(t, ok)

				assert.Equal(t, a.userID, claims.UserID)
				assert.Equal(t, a.displayName, claims.DisplayName)
				assert.Equal(t, a.email, claims.Email)

				assert.Equal(t, f.issuer, claims.Issuer)
				assert.Contains(t, claims.Audience, f.audience)

				assert.NotNil(t, claims.ExpiresAt)
				assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			keyPath := createTempPrivateKeyFile(t)

			generator, err := NewTokenGenerator(
				keyPath,
				tc.fields.issuer,
				tc.fields.audience,
				tc.fields.expiry,
			)
			require.NoError(t, err)

			tokenString, err := generator.GenerateToken(
				tc.args.userID,
				tc.args.displayName,
				tc.args.email,
			)

			tc.verify(t, tokenString, err, tc.args, tc.fields)
		})
	}
}