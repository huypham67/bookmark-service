package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/huypham67/bookmark-service/internal/dto/request"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/repository"
	"github.com/huypham67/bookmark-service/pkg/jwtutils"
	"github.com/huypham67/bookmark-service/pkg/security"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrUsernameAlreadyExists  = errors.New("username already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidPassword        = errors.New("invalid password")
)

// User defines the contract for user services.
type User interface {
	RegisterUser(ctx context.Context, req request.RegisterUserRequest) (*model.User, error)
	LoginUser(ctx context.Context, req request.LoginRequest) (string, error)
	GetUserInfo(ctx context.Context, userID string) (*model.User, error)
}

type userService struct {
	userRepo       repository.User
	passwordHasher security.PasswordHasher
	tokenGenerator jwtutils.TokenGenerator
}

// NewUserService creates a new user service with the given user repository, password hasher, and token generator.
func NewUserService(userRepo repository.User, passwordHasher security.PasswordHasher, tokenGenerator jwtutils.TokenGenerator) User {
	return &userService{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
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
		return nil, ErrEmailAlreadyRegistered
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
		return nil, ErrUsernameAlreadyExists
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

// LoginUser authenticates a user by validating credentials and returns a JWT token.
func (s *userService) LoginUser(ctx context.Context, req request.LoginRequest) (string, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().
			Err(err).
			Str("username", req.Username).
			Msg("failed to get user by username")
		return "", err
	}

	if user == nil {
		log.Warn().
			Str("username", req.Username).
			Msg("user not found")
		return "", ErrUserNotFound
	}

	// Validate password
	if err := s.passwordHasher.Compare(user.Password, req.Password); err != nil {
		log.Warn().
			Str("username", req.Username).
			Msg("invalid password")
		return "", ErrInvalidPassword
	}

	// Generate JWT token
	token, err := s.tokenGenerator.GenerateToken(user.ID, user.DisplayName, user.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", user.ID).
			Msg("failed to generate token")
		return "", err
	}

	log.Info().
		Str("user_id", user.ID).
		Str("username", user.Username).
		Msg("user logged in successfully")

	return token, nil
}

// GetUserInfo retrieves user information by user ID.
func (s *userService) GetUserInfo(ctx context.Context, userID string) (*model.User, error) {
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
