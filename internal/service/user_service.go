package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/huypham67/bookmark-service/internal/dto/request"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/repository"
	"github.com/huypham67/bookmark-service/pkg/security"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// User defines the contract for user services.
type User interface {
	RegisterUser(ctx context.Context, req request.RegisterUserRequest) (*model.User, error)
}

type userService struct {
	userRepo       repository.User
	passwordHasher security.PasswordHasher
}

// NewUserService creates a new user service with the given user repository and password hasher.
func NewUserService(userRepo repository.User, passwordHasher security.PasswordHasher) User {
	return &userService{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}
}

// RegisterUser registers a new user by validating input, hashing password, and saving to database.
func (s *userService) RegisterUser(ctx context.Context, req request.RegisterUserRequest) (*model.User, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("failed to check if email exists")
		return nil, err
	}

	if existingUser != nil {
		log.Warn().
			Str("email", req.Email).
			Msg("email already registered")
		return nil, errors.New("email already registered")
	}

	// Check if username already exists
	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().
			Err(err).
			Str("username", req.Username).
			Msg("failed to check if username exists")
		return nil, err
	}

	if existingUser != nil {
		log.Warn().
			Str("username", req.Username).
			Msg("username already exists")
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to hash password")
		return nil, err
	}

	// Create new user
	user := &model.User{
		ID:          uuid.New().String(),
		DisplayName: req.DisplayName,
		Username:    req.Username,
		Email:       req.Email,
		Password:    hashedPassword,
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("username", req.Username).
			Msg("failed to register user")
		return nil, err
	}

	log.Info().
		Str("user_id", user.ID).
		Str("email", user.Email).
		Str("username", user.Username).
		Msg("user registered successfully")

	return user, nil
}
