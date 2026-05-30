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

// Profile defines the contract for user profile HTTP handlers.
type Profile interface {
	GetUserInfo(c *gin.Context)
	UpdateUserInfo(c *gin.Context)
}

type profileHandler struct {
	profileService service.Profile
}

// NewProfileHandler creates a new profile handler with the given profile service.
func NewProfileHandler(profileService service.Profile) Profile {
	return &profileHandler{
		profileService: profileService,
	}
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
func (h *profileHandler) GetUserInfo(c *gin.Context) {
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

	user, err := h.profileService.GetUserInfo(c, userID)

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

// UpdateUserInfo handles the user info update endpoint.
//
// @Summary Update User Info
// @Description Update authenticated user's display name and email
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body request.UpdateUserRequest true "User update data"
// @Success 200 {object} response.UpdateUserResponse "User updated successfully"
// @Failure 400 {object} gin.H "Invalid request body"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 409 {object} gin.H "Email already exists"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/self/info [put]
func (h *profileHandler) UpdateUserInfo(c *gin.Context) {
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

	var req request.UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	userID := claims.UserID

	if err := h.profileService.UpdateUserInfo(c, userID, req); err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Str("email", req.Email).
			Msg("failed to update user info")

		// Check if email already exists
		if errors.Is(err, service.ErrEmailAlreadyRegistered) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})

		return
	}

	c.JSON(http.StatusOK, response.UpdateUserResponse{
		Message: "Edit current user successfully!",
	})
}
