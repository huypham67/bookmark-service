package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel defines the common fields for all database models.
type BaseModel struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid;column:id"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime:milli;column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime:milli;column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate is a GORM hook that is called before a new record is created in the database.
// It generates a new UUID for the ID field if it is not already set.
func (m *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}

	return nil
}
