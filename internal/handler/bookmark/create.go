package bookmark

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	"github.com/huypham67/bookmark-common/pkg/requestutils"
	"github.com/huypham67/bookmark-common/pkg/response"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/rs/zerolog/log"
)

// Create handles the bookmark creation endpoint.
//
// @Summary Create Bookmark
// @Description Create a new bookmark for the authenticated user
// @Tags bookmarks
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body bookmarkDTO.CreateBookmarkRequest true "Bookmark creation data"
// @Success 201 {object} bookmarkDTO.BookmarkResponse "Bookmark created successfully"
// @Failure 400 {object} gin.H "Invalid request body"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 409 {object} gin.H "Bookmark code already exists"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/bookmarks [post]
func (h *handler) Create(c *gin.Context) {
	userID, err := jwt.GetUserIDFromContext(c)

	if err != nil {
		log.Warn().Msg("user ID not found in context")
		response.Unauthorized(c, "Unauthorized")
		return
	}

	req, err := requestutils.Bind[bookmarkDTO.CreateBookmarkRequest](c)

	if err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	bm, err := h.service.Create(c, userID, *req)

	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Str("url", req.URL).
			Msg("failed to create bookmark")

		switch {
		case errors.Is(err, bookmark.ErrBookmarkAlreadyExists):
			response.Conflict(c, "Bookmark code already exists")
		case errors.Is(err, bookmark.ErrBadRequest):
			response.BadRequest(c, "Invalid bookmark request")
		default:
			response.InternalServerError(c)
		}
		return
	}

	c.JSON(http.StatusCreated, response.Success(
		bookmarkDTO.BookmarkData{
			ID:          bm.ID,
			Code:        bm.Code,
			Description: bm.Description,
			URL:         bm.URL,
			CreatedAt:   bm.CreatedAt,
			UpdatedAt:   bm.UpdatedAt,
		},
		"Bookmark created successfully!",
	))
}
