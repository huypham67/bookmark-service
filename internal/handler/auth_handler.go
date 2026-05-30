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

// Auth defines the contract for authentication HTTP handlers.
type Auth interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type authHandler struct {
	authService service.Auth
}

// NewAuthHandler creates a new auth handler with the given auth service.
func NewAuthHandler(authService service.Auth) Auth {
	return &authHandler{
		authService: authService,
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
func (h *authHandler) Register(c *gin.Context) {
	var req request.RegisterUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	user, err := h.authService.RegisterUser(c, req)

	if err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("username", req.Username).
			Msg("failed to register user")

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

// Login handles the user login endpoint.
//
// @Summary Login User
// @Description Login a user with username and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "User login data"
// @Success 200 {object} response.LoginResponse "User logged in successfully"
// @Failure 400 {object} gin.H "Invalid request body"
// @Failure 401 {object} gin.H "Invalid credentials"
// @Failure 404 {object} gin.H "User not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/users/login [post]
func (h *authHandler) Login(c *gin.Context) {
	var req request.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	token, err := h.authService.LoginUser(c, req)

	if err != nil {
		log.Error().
			Err(err).
			Str("username", req.Username).
			Msg("failed to login user")

		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInvalidPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid username or password",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})

		return
	}

	c.JSON(http.StatusOK, response.LoginResponse{
		Data:    token,
		Message: "Logged in successfully!",
	})
}
