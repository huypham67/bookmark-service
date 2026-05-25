package security

import "golang.org/x/crypto/bcrypt"

// PasswordHasher defines the contract for password hashing operations.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type bcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher creates a new bcrypt-based password hasher.
func NewBcryptPasswordHasher() PasswordHasher {
	return &bcryptPasswordHasher{
		cost: bcrypt.DefaultCost,
	}
}

// Hash generates a bcrypt hash from the given password string.
func (h *bcryptPasswordHasher) Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Compare compares a bcrypt hashed password with its possible plaintext equivalent.
func (h *bcryptPasswordHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
