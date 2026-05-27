package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-service/internal/dto/request"
	"github.com/huypham67/bookmark-service/internal/dto/response"
	"github.com/huypham67/bookmark-service/internal/service"
	"github.com/rs/zerolog/log"
)

// User defines the contract for user HTTP handlers.
type User interface {
	Register(c *gin.Context)
}

type userHandler struct {
	userService service.User
}

// NewUserHandler creates a new user handler with the given user service.
func NewUserHandler(userService service.User) User {
	return &userHandler{
		userService: userService,
	}
}

// Register handles the user registration endpoint.
//
// @Summary Register User
// @Description Register a new user with email, username, and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body request.RegisterUserRequest true "User registration data"
// @Success 201 {object} response.RegisterUserResponse "User registered successfully"
// @Failure 400 {object} gin.H "Invalid request body"
// @Failure 409 {object} gin.H "User already exists"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/users/register [post]
func (h *userHandler) Register(c *gin.Context) {
	var req request.RegisterUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	user, err := h.userService.RegisterUser(c, req)

	if err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("username", req.Username).
			Msg("500 - failed to register user")

		// Check if it's a validation error (email/username exists)
		if errors.Is(err, service.ErrEmailAlreadyRegistered) || errors.Is(err, service.ErrUsernameAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "User already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})

		return
	}

	c.JSON(http.StatusCreated, response.RegisterUserResponse{
		Data: response.UserData{
			ID:          user.ID,
			DisplayName: user.DisplayName,
			Username:    user.Username,
			Email:       user.Email,
			CreatedAt:   user.CreatedAt,
		},
		Message: "Register an user successfully!",
	})
}
