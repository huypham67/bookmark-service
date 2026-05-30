package repository

import (
	"context"

	"github.com/huypham67/bookmark-service/internal/model"
	"gorm.io/gorm"
)

// User defines the contract for user repository operations.
// mockery --name=User --dir=internal/repository --output=internal/repository/mocks --filename=user_repository.go
type User interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, userID string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository with the given GORM database client.
func NewUserRepository(db *gorm.DB) User {
	return &userRepository{
		db: db,
	}
}

// Create saves a new user to the database.
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByEmail retrieves a user by their email address.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user *model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetByUsername retrieves a user by their username.
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user *model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetByID retrieves a user by their ID.
func (r *userRepository) GetByID(ctx context.Context, userID string) (*model.User, error) {
	var user *model.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates user's display_name and email.
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"display_name": user.DisplayName,
		"email":        user.Email,
	}).Error
}
