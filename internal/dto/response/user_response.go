package response

import "time"

// UserData represents the user data in the response.
type UserData struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
}

// RegisterUserResponse represents the user registration response payload.
type RegisterUserResponse struct {
	Data    UserData `json:"data"`
	Message string   `json:"message"`
}
