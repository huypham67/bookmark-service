package service

import (
	"context"

	"github.com/huypham67/bookmark-service/internal/dto/request"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/repository"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Profile defines the contract for user profile services.
// mockery --name=Profile --dir=internal/service --output=internal/service/mocks --filename=profile_service.go
type Profile interface {
	GetUserInfo(ctx context.Context, userID string) (*model.User, error)
	UpdateUserInfo(ctx context.Context, userID string, req request.UpdateUserRequest) error
}

type profileService struct {
	userRepo repository.User
}

// NewProfileService creates a new profile service with the given user repository.
func NewProfileService(userRepo repository.User) Profile {
	return &profileService{
		userRepo: userRepo,
	}
}

// GetUserInfo retrieves user information by user ID.
func (s *profileService) GetUserInfo(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Msg("failed to get user by ID")
		return nil, err
	}

	return user, nil
}

// UpdateUserInfo updates user's display name and email.
func (s *profileService) UpdateUserInfo(ctx context.Context, userID string, req request.UpdateUserRequest) error {
	// Check if the new email already exists for a different user
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("failed to check if email exists")
		return err
	}

	// If email exists and belongs to a different user, return error
	if existingUser != nil && existingUser.ID != userID {
		log.Warn().
			Str("email", req.Email).
			Str("user_id", userID).
			Msg("email already registered to another user")
		return ErrEmailAlreadyRegistered
	}

	// Update user
	user := &model.User{
		ID:          userID,
		DisplayName: req.DisplayName,
		Email:       req.Email,
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Msg("failed to update user")
		return err
	}

	log.Info().
		Str("user_id", userID).
		Str("email", req.Email).
		Msg("user updated successfully")

	return nil
}
