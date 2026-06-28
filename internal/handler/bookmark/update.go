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
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog/log"
)

// Update handles the bookmark update endpoint.
//
// @Summary Update Bookmark
// @Description Update an existing bookmark (description and/or URL)
// @Tags bookmarks
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Bookmark ID"
// @Param request body object{description=string,url=string} false "Updated bookmark data"
// @Success 200 {object} gin.H "Bookmark updated successfully"
// @Failure 400 {object} gin.H "Invalid bookmark ID or request data"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 404 {object} gin.H "Bookmark not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/bookmarks/{id} [put]
func (h *handler) Update(c *gin.Context) {
	segment := newrelic.FromContext(c).StartSegment("handler.bookmark.Update")
	defer segment.End()

	userID, err := jwt.GetUserIDFromContext(c)

	if err != nil {
		log.Warn().Msg("user ID not found in context")
		response.Unauthorized(c, "Unauthorized")
		return
	}

	req, err := requestutils.Bind[bookmarkDTO.UpdateBookmarkRequest](c)

	if err != nil {
		log.Warn().
			Err(err).
			Msg("invalid update bookmark request")
		response.BadRequest(c, "Invalid request")
		return
	}

	if err := h.service.Update(c, userID, req.ID, *req); err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Str("bookmark_id", req.ID).
			Msg("failed to update bookmark")

		switch {
		case errors.Is(err, bookmark.ErrBookmarkNotFound):
			response.NotFound(c, "Bookmark not found")
		case errors.Is(err, bookmark.ErrBadRequest):
			response.BadRequest(c, "Invalid request")
		default:
			response.InternalServerError(c)
		}
		return
	}

	c.JSON(http.StatusOK, response.Message("Success"))
}
