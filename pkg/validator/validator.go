package validator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	instance *validator.Validate
	once     sync.Once
)

// Get returns the singleton validator instance
func Get() *validator.Validate {
	once.Do(func() {
		instance = validator.New()
	})
	return instance
}
