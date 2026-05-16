package utils

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// CodeGenerator defines the contract for generating random codes.
type CodeGenerator interface {
	Generate(length int) (string, error)
}

type codeGenerator struct{}

// NewCodeGenerator creates a new code generator instance.
func NewCodeGenerator() CodeGenerator {
	return &codeGenerator{}
}

// Generate creates a random string of the specified length using characters from the defined charset.
func (cg *codeGenerator) Generate(length int) (string, error) {
	result := make([]byte, length)

	for i := range result {
		randomIndex, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(len(charset))),
		)

		if err != nil {
			return "", err
		}

		result[i] = charset[randomIndex.Int64()]
	}

	return string(result), nil
}
