package service

import (
	"context"
	"testing"

	"github.com/huypham67/bookmark-service/internal/dto/request"
	"github.com/huypham67/bookmark-service/internal/model"
	repositoryMocks "github.com/huypham67/bookmark-service/internal/repository/mocks"
	securityMocks "github.com/huypham67/bookmark-service/pkg/security/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func expectedRegisteredUser() *model.User {
	return &model.User{
		DisplayName: "Test Display Name",
		Username:    "testuser",
		Email:       "testuser@gmail.com",
		Password:    "$2a$10$hashedpassword123456789",
	}
}

func matchUser(expected *model.User) interface{} {
	return mock.MatchedBy(func(actual *model.User) bool {
		return actual.DisplayName == expected.DisplayName &&
			actual.Username == expected.Username &&
			actual.Email == expected.Email &&
			actual.Password == expected.Password &&
			actual.ID != ""
	})
}

func TestUserService_RegisterUser(t *testing.T) {
	t.Parallel()

	type args struct {
		request request.RegisterUserRequest
	}

	testCases := []struct {
		name           string
		args           args
		setupMocks     func(context.Context, *repositoryMocks.User, *securityMocks.PasswordHasher)
		verifyResponse func(*testing.T, *model.User, error)
	}{
		{
			name: "should register user successfully",
			args: args{
				request: request.RegisterUserRequest{
					DisplayName: "Test Display Name",
					Username:    "testuser",
					Email:       "testuser@gmail.com",
					Password:    "password123",
				},
			},
			setupMocks: func(
				ctx context.Context,
				userRepo *repositoryMocks.User,
				passwordHasher *securityMocks.PasswordHasher,
			) {
				passwordHasher.
					On("Hash", "password123").
					Return("$2a$10$hashedpassword123456789", nil).
					Once()

				expectedUser := expectedRegisteredUser()

				userRepo.
					On(
						"GetByEmail",
						ctx,
						expectedUser.Email,
					).
					Return(nil, gorm.ErrRecordNotFound).
					Once()

				userRepo.
					On(
						"GetByUsername",
						ctx,
						expectedUser.Username,
					).
					Return(nil, gorm.ErrRecordNotFound).
					Once()

				userRepo.
					On(
						"Create",
						ctx,
						matchUser(expectedUser),
					).
					Return(nil).
					Once()
			},
			verifyResponse: func(
				t *testing.T,
				user *model.User,
				err error,
			) {
				assert.NoError(t, err)
				require.NotNil(t, user)

				assert.Equal(t, "Test Display Name", user.DisplayName)
				assert.Equal(t, "testuser", user.Username)
				assert.Equal(t, "testuser@gmail.com", user.Email)
				assert.Equal(t, "$2a$10$hashedpassword123456789", user.Password)
				assert.NotEmpty(t, user.ID)
			},
		},
		{
			name: "should return error when user already exists",
			args: args{
				request: request.RegisterUserRequest{
					DisplayName: "Test Display Name",
					Username:    "testuser",
					Email:       "testuser@gmail.com",
					Password:    "password123",
				},
			},
			setupMocks: func(
				ctx context.Context,
				userRepo *repositoryMocks.User,
				passwordHasher *securityMocks.PasswordHasher,
			) {
				userRepo.
					On(
						"GetByEmail",
						ctx,
						"testuser@gmail.com",
					).
					Return(&model.User{}, nil).
					Once()
			},
			verifyResponse: func(
				t *testing.T,
				user *model.User,
				err error,
			) {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.ErrorIs(t, err, ErrEmailAlreadyRegistered)
			},
		},
		{
			name: "should return error when checking email fails",
			args: args{
				request: request.RegisterUserRequest{
					DisplayName: "Test Display Name",
					Username:    "testuser",
					Email:       "testuser@gmail.com",
					Password:    "password123",
				},
			},
			setupMocks: func(
				ctx context.Context,
				userRepo *repositoryMocks.User,
				passwordHasher *securityMocks.PasswordHasher,
			) {
				userRepo.
					On(
						"GetByEmail",
						ctx,
						"testuser@gmail.com",
					).
					Return(nil, assert.AnError).
					Once()
			},
			verifyResponse: func(
				t *testing.T,
				user *model.User,
				err error,
			) {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.EqualError(t, err, assert.AnError.Error())
			},
		},
		{
			name: "should return error when username already exists",
			args: args{
				request: request.RegisterUserRequest{
					DisplayName: "Test Display Name",
					Username:    "testuser",
					Email:       "testuser@gmail.com",
					Password:    "password123",
				},
			},
			setupMocks: func(
				ctx context.Context,
				userRepo *repositoryMocks.User,
				passwordHasher *securityMocks.PasswordHasher,
			) {
				userRepo.
					On(
						"GetByEmail",
						ctx,
						"testuser@gmail.com",
					).Return(nil, gorm.ErrRecordNotFound).
					Once()

				userRepo.
					On(
						"GetByUsername",
						ctx,
						"testuser",
					).
					Return(&model.User{}, nil).
					Once()
			},
			verifyResponse: func(
				t *testing.T,
				user *model.User,
				err error,
			) {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.ErrorIs(t, err, ErrUsernameAlreadyExists)
			},
		},
		{
			name: "should return error when checking username fails",
			args: args{
				request: request.RegisterUserRequest{
					DisplayName: "Test Display Name",
					Username:    "testuser",
					Email:       "testuser@gmail.com",
					Password:    "password123",
				},
			},
			setupMocks: func(
				ctx context.Context,
				userRepo *repositoryMocks.User,
				passwordHasher *securityMocks.PasswordHasher,
			) {
				userRepo.
					On(
						"GetByEmail",
						ctx,
						"testuser@gmail.com",
					).Return(nil, gorm.ErrRecordNotFound).
					Once()

				userRepo.
					On(
						"GetByUsername",
						ctx,
						"testuser",
					).Return(nil, assert.AnError).
					Once()
			},
			verifyResponse: func(
				t *testing.T,
				user *model.User,
				err error,
			) {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.EqualError(t, err, assert.AnError.Error())
			},
		},
		{
			name: "should return error when password hashing fails",
			args: args{
				request: request.RegisterUserRequest{
					DisplayName: "Test Display Name",
					Username:    "testuser",
					Email:       "testuser@gmail.com",
					Password:    "password123",
				},
			},
			setupMocks: func(
				ctx context.Context,
				userRepo *repositoryMocks.User,
				passwordHasher *securityMocks.PasswordHasher,
			) {
				passwordHasher.
					On("Hash", "password123").
					Return("", assert.AnError).
					Once()

				userRepo.
					On(
						"GetByEmail",
						ctx,
						"testuser@gmail.com",
					).Return(nil, gorm.ErrRecordNotFound).
					Once()

				userRepo.
					On(
						"GetByUsername",
						ctx,
						"testuser",
					).Return(nil, gorm.ErrRecordNotFound).
					Once()
			},
			verifyResponse: func(
				t *testing.T,
				user *model.User,
				err error,
			) {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.EqualError(t, err, assert.AnError.Error())
			},
		},
		{
			name: "should return error when creating user in database fails",
			args: args{
				request: request.RegisterUserRequest{
					DisplayName: "Test Display Name",
					Username:    "testuser",
					Email:       "testuser@gmail.com",
					Password:    "password123",
				},
			},
			setupMocks: func(
				ctx context.Context,
				userRepo *repositoryMocks.User,
				passwordHasher *securityMocks.PasswordHasher,
			) {
				passwordHasher.
					On("Hash", "password123").
					Return("$2a$10$hashedpassword123456789", nil).
					Once()

				expectedUser := expectedRegisteredUser()

				userRepo.
					On(
						"GetByEmail",
						ctx,
						expectedUser.Email,
					).
					Return(nil, gorm.ErrRecordNotFound).
					Once()

				userRepo.
					On(
						"GetByUsername",
						ctx,
						expectedUser.Username,
					).
					Return(nil, gorm.ErrRecordNotFound).
					Once()

				userRepo.
					On(
						"Create",
						ctx,
						matchUser(expectedUser),
					).
					Return(assert.AnError).
					Once()
			},
			verifyResponse: func(
				t *testing.T,
				user *model.User,
				err error,
			) {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.EqualError(t, err, assert.AnError.Error())
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			userRepo := new(repositoryMocks.User)
			passwordHasher := securityMocks.NewPasswordHasher(t)

			tc.setupMocks(ctx, userRepo, passwordHasher)

			userService := NewUserService(userRepo, passwordHasher)

			user, err := userService.RegisterUser(
				ctx,
				tc.args.request,
			)

			tc.verifyResponse(t, user, err)

			userRepo.AssertExpectations(t)
		})
	}
}
