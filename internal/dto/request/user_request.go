package request

// RegisterUserRequest represents the user registration request payload.
type RegisterUserRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
	Username    string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
}
