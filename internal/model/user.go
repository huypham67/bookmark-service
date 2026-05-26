package model

import "time"

// User represents a user entity stored in the database.
type User struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;column:id"`
	DisplayName string    `json:"display_name" gorm:"not null;column:display_name"`
	Username    string    `json:"username" gorm:"not null;unique;column:username"`
	Email       string    `json:"email" gorm:"unique;not null;column:email"`
	Password    string    `json:"-" gorm:"not null;column:password"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime:milli;column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime:milli;column:updated_at"`
}
