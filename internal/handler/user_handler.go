package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-service/internal/dto/request"
	"github.com/huypham67/bookmark-service/internal/dto/response"
	"github.com/huypham67/bookmark-service/internal/service"
	"github.com/huypham67/bookmark-service/pkg/jwtutils"
	"github.com/rs/zerolog/log"
)

// User defines the contract for user HTTP handlers.
type User interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	GetUserInfo(c *gin.Context)
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
func (h *userHandler) Login(c *gin.Context) {
	var req request.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	token, err := h.userService.LoginUser(c, req)

	if err != nil {
		log.Error().
			Err(err).
			Str("username", req.Username).
			Msg("failed to login user")

		// Check error types
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		if errors.Is(err, service.ErrInvalidPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials",
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

// GetUserInfo handles the user info endpoint.
//
// @Summary Get User Info
// @Description Get authenticated user information from JWT token
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} response.UserResponse "User information"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 404 {object} gin.H "User not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/self/info [get]
func (h *userHandler) GetUserInfo(c *gin.Context) {
	// Extract claims from context (set by JWT middleware)
	claimsObj, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "missing claims in context",
		})
		return
	}

	claims, ok := claimsObj.(*jwtutils.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid claims type",
		})
		return
	}

	userID := claims.UserID

	user, err := h.userService.GetUserInfo(c, userID)

	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Msg("failed to get user info")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})

		return
	}

	c.JSON(http.StatusOK, response.UserResponse{
		Data:    user,
		Message: "User information retrieved successfully!",
	})
}
