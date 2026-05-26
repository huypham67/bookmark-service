package jwtutils

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRSATokenValidator_ValidateToken(t *testing.T) {
	t.Parallel()

	type fields struct {
		issuer   string
		audience string
		expiry   time.Duration
	}

	type args struct {
		tokenIssuer   string
		tokenAudience string
		tokenString   string
	}

	testCases := []struct {
		name   string
		fields fields
		args   args
		verify func(*testing.T, *CustomClaims, error)
	}{
		{
			name: "should return claims when token is valid",
			fields: fields{
				issuer:   "bookmark-service",
				audience: "bookmark-client",
				expiry:   time.Hour,
			},
			args: args{
				tokenIssuer:   "bookmark-service",
				tokenAudience: "bookmark-client",
			},
			verify: func(t *testing.T, claims *CustomClaims, err error) {
				require.NoError(t, err)
				require.NotNil(t, claims)

				assert.Equal(t, "user-123", claims.UserID)
				assert.Equal(t, "John Doe", claims.DisplayName)
				assert.Equal(t, "john@example.com", claims.Email)
			},
		},
		{
			name: "should return error when issuer is invalid",
			fields: fields{
				issuer:   "bookmark-service",
				audience: "bookmark-client",
				expiry:   time.Hour,
			},
			args: args{
				tokenIssuer:   "wrong-issuer",
				tokenAudience: "bookmark-client",
			},
			verify: func(t *testing.T, claims *CustomClaims, err error) {
				require.Error(t, err)
				assert.Nil(t, claims)
				assert.Contains(t, err.Error(), "invalid token issuer")
			},
		},
		{
			name: "should return error when audience is invalid",
			fields: fields{
				issuer:   "bookmark-service",
				audience: "bookmark-client",
				expiry:   time.Hour,
			},
			args: args{
				tokenIssuer:   "bookmark-service",
				tokenAudience: "wrong-audience",
			},
			verify: func(t *testing.T, claims *CustomClaims, err error) {
				require.Error(t, err)
				assert.Nil(t, claims)
				assert.Contains(t, err.Error(), "invalid token audience")
			},
		},
		{
			name: "should return error when token is malformed",
			fields: fields{
				issuer:   "bookmark-service",
				audience: "bookmark-client",
				expiry:   time.Hour,
			},
			args: args{
				tokenString: "invalid-token",
			},
			verify: func(t *testing.T, claims *CustomClaims, err error) {
				require.Error(t, err)
				assert.Nil(t, claims)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)
			publicKey := &privateKey.PublicKey

			validator, err := NewTokenValidator(
				publicKey,
				tc.fields.issuer,
				tc.fields.audience,
			)
			require.NoError(t, err)

			tokenString := tc.args.tokenString

			if tokenString == "" {
				generator, err := NewTokenGenerator(
					privateKey,
					tc.args.tokenIssuer,
					tc.args.tokenAudience,
					tc.fields.expiry,
				)
				require.NoError(t, err)

				tokenString, err = generator.GenerateToken(
					"user-123",
					"John Doe",
					"john@example.com",
				)
				require.NoError(t, err)
			}

			claims, err := validator.ValidateToken(tokenString)

			tc.verify(t, claims, err)
		})
	}
}
