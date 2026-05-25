package testutil

import "github.com/huypham67/bookmark-service/internal/model"

func NewUserFixtures() []model.User {
	return []model.User{
		{
			ID:          "abcd-1234-efgh-5678",
			DisplayName: "Test User 1",
			Username:    "testuser1",
			Email:       "testuser1@gmail.com",
			Password:    "hashed_password_1",
		},
		{
			ID:          "abcde-5678-fghi-1234",
			DisplayName: "Test User 2",
			Username:    "testuser2",
			Email:       "testuser2@gmail.com",
			Password:    "hashed_password_2",
		},
		{
			ID:          "s876-4321-hijk-5678",
			DisplayName: "Test User 3",
			Username:    "testuser3",
			Email:       "testuser3@gmail.com",
			Password:    "hashed_password_3",
		},
	}
}