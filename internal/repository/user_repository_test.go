package repository

import (
	"context"
	"testing"

	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/repository/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserRepository_Create(t *testing.T) {
	t.Parallel()

	type args struct {
		user model.User
	}

	testCases := []struct {
		name   string
		args   args
		verify func(*testing.T, *gorm.DB, error, args)
	}{
		{
			name: "should create user successfully",
			args: args{
				user: model.User{
					ID:          "asdfg-5678-hijk-1234",
					DisplayName: "Test User 4",
					Username:    "testuser4",
					Email:       "testuser4@gmail.com",
					Password:    "hashed-password",
				},
			},
			verify: func(t *testing.T, db *gorm.DB, err error, a args) {
				require.NoError(t, err)

				var actual model.User
				err = db.First(&actual, "id = ?", a.user.ID).Error
				require.NoError(t, err)

				assert.Equal(t, a.user.ID, actual.ID)
				assert.Equal(t, a.user.DisplayName, actual.DisplayName)
				assert.Equal(t, a.user.Username, actual.Username)
				assert.Equal(t, a.user.Email, actual.Email)
				assert.Equal(t, a.user.Password, actual.Password)
			},
		},
		{
			name: "should return error when email already exists",
			args: args{
				user: model.User{
					ID:          "zxcvb-5678-asdf-1234",
					DisplayName: "Test User 5",
					Username:    "testuser5",
					Email:       "testuser1@gmail.com", // duplicate email
					Password:    "hashed-password",
				},
			},
			verify: func(t *testing.T, db *gorm.DB, err error, a args) {
				require.Error(t, err)

				var actual model.User
				err = db.First(&actual, "id = ?", a.user.ID).Error
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			},
		},
		{
			name: "should return error when username already exists",
			args: args{
				user: model.User{
					ID:          "qwert-5678-zxcv-1234",
					DisplayName: "Test User 6",
					Username:    "testuser1", // duplicate username
					Email:       "testuser6@gmail.com",
					Password:    "hashed-password",
				},
			},
			verify: func(t *testing.T, db *gorm.DB, err error, a args) {
				require.Error(t, err)

				var actual model.User
				err = db.First(&actual, "id = ?", a.user.ID).Error
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			testDB := testutil.SetupTestDatabase(t)
			repo := NewUserRepository(testDB)

			err := repo.Create(ctx, &tc.args.user)

			tc.verify(t, testDB, err, tc.args)
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Parallel()

	type args struct {
		email string
	}

	testCases := []struct {
		name   string
		args   args
		verify func(*testing.T, *model.User, error)
	}{
		{
			name: "should return user when email exists",
			args: args{
				email: "testuser1@gmail.com",
			},
			verify: func(t *testing.T, user *model.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, "abcd-1234-efgh-5678", user.ID)
				assert.Equal(t, "Test User 1", user.DisplayName)
				assert.Equal(t, "testuser1", user.Username)
				assert.Equal(t, "testuser1@gmail.com", user.Email)
				assert.Equal(t, "hashed_password_1", user.Password)
			},
		},
		{
			name: "should return nil when email does not exist",
			args: args{
				email: "nonexistent@gmail.com",
			},
			verify: func(t *testing.T, user *model.User, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
				assert.Nil(t, user)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			testDB := testutil.SetupTestDatabase(t)
			repo := NewUserRepository(testDB)

			user, err := repo.GetByEmail(ctx, tc.args.email)

			tc.verify(t, user, err)
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	t.Parallel()

	type args struct {
		username string
	}

	testCases := []struct {
		name   string
		args   args
		verify func(*testing.T, *model.User, error)
	}{
		{
			name: "should return user when username exists",
			args: args{
				username: "testuser1",
			},
			verify: func(t *testing.T, user *model.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, "abcd-1234-efgh-5678", user.ID)
				assert.Equal(t, "Test User 1", user.DisplayName)
				assert.Equal(t, "testuser1", user.Username)
				assert.Equal(t, "testuser1@gmail.com", user.Email)
				assert.Equal(t, "hashed_password_1", user.Password)
			},
		},
		{
			name: "should return nil when username does not exist",
			args: args{
				username: "nonexistent",
			},
			verify: func(t *testing.T, user *model.User, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
				assert.Nil(t, user)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			testDB := testutil.SetupTestDatabase(t)
			repo := NewUserRepository(testDB)

			user, err := repo.GetByUsername(ctx, tc.args.username)

			tc.verify(t, user, err)
		})
	}
}
