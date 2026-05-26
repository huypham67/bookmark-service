package main

import (
	"fmt"
	"log"
	"time"

	"github.com/huypham67/bookmark-service/pkg/jwtutils"
)

func main() {
	// Use existing keys from /keys directory
	privateKeyFile := "keys/private.pem"
	publicKeyFile := "keys/public.pem"

	// Create token generator
	issuer := "bookmark-service"
	audience := "bookmark-api"
	expiry := 24 * time.Hour

	generator, err := jwtutils.NewTokenGenerator(privateKeyFile, issuer, audience, expiry)
	if err != nil {
		log.Fatalf("Failed to create token generator: %v", err)
	}

	// Generate token
	userID := "12345"
	displayName := "Pham Huy"
	email := "phamhuy@gmail.com"

	token, err := generator.GenerateToken(userID, displayName, email)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}
	fmt.Printf("\n✓ Token generated successfully:\n%s\n", token)

	// Create token validator
	validator, err := jwtutils.NewTokenValidator(publicKeyFile, issuer, audience)
	if err != nil {
		log.Fatalf("Failed to create token validator: %v", err)
	}

	// Validate token
	claims, err := validator.ValidateToken(token)
	if err != nil {
		log.Fatalf("Failed to validate token: %v", err)
	}

	fmt.Printf("\n✓ Token validated successfully!\n")
	fmt.Printf("  UserID: %s\n", claims.UserID)
	fmt.Printf("  DisplayName: %s\n", claims.DisplayName)
	fmt.Printf("  Email: %s\n", claims.Email)
	fmt.Printf("  Issuer: %s\n", claims.Issuer)
	fmt.Printf("  Audience: %v\n", claims.Audience)
	fmt.Printf("  ExpiresAt: %v\n", claims.ExpiresAt)

	// Test with invalid token
	fmt.Printf("\n--- Testing with invalid token ---\n")
	_, err = validator.ValidateToken("invalid.token.here")
	if err != nil {
		fmt.Printf("✓ Invalid token correctly rejected: %v\n", err)
	}

	fmt.Printf("\n✓ Test completed successfully!\n")
}
