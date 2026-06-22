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

// Delete handles the bookmark deletion endpoint.
//
// @Summary Delete Bookmark
// @Description Delete an existing bookmark
// @Tags bookmarks
// @Produce json
// @Security Bearer
// @Param id path string true "Bookmark ID"
// @Success 200 {object} gin.H "Bookmark deleted successfully"
// @Failure 400 {object} gin.H "Invalid bookmark ID"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 404 {object} gin.H "Bookmark not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/bookmarks/{id} [delete]
func (h *handler) Delete(c *gin.Context) {
	segment := newrelic.FromContext(c).StartSegment("handler.bookmark.Delete")
	defer segment.End()

	userID, err := jwt.GetUserIDFromContext(c)

	if err != nil {
		log.Warn().Msg("user ID not found in context")
		response.Unauthorized(c, "Unauthorized")
		return
	}

	req, err := requestutils.Bind[bookmarkDTO.DeleteBookmarkRequest](c)

	if err != nil {
		log.Warn().
			Err(err).
			Msg("invalid delete bookmark request")
		response.BadRequest(c, "invalid request")
		return
	}

	if err := h.service.Delete(c, userID, req.ID); err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Str("bookmark_id", req.ID).
			Msg("failed to delete bookmark")

		switch {
		case errors.Is(err, bookmark.ErrBookmarkNotFound):
			response.NotFound(c, "Bookmark not found")
		case errors.Is(err, bookmark.ErrBadRequest):
			response.BadRequest(c, "invalid request")
		default:
			response.InternalServerError(c)
		}
		return
	}

	c.JSON(http.StatusOK, response.Message("Success"))
}
