package bookmark

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	"github.com/huypham67/bookmark-common/pkg/requestutils"
	"github.com/huypham67/bookmark-common/pkg/response"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/rs/zerolog/log"
)

// List handles the bookmark list retrieval endpoint with pagination.
//
// @Summary List Bookmarks
// @Description Get paginated list of bookmarks for the authenticated user
// @Tags bookmarks
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number (default: 1)" default(1)
// @Param limit query int false "Items per page, max 100 (default: 10)" default(10)
// @Param sort query string false "Sort field: created_at, updated_at, code, url (default: created_at)" Enums(created_at,updated_at,code,url) default(created_at)
// @Success 200 {object} bookmarkDTO.BookmarkListResponse "List of bookmarks with pagination"
// @Failure 400 {object} gin.H "Invalid query parameters"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/bookmarks [get]
func (h *handler) List(c *gin.Context) {
	userID, err := jwt.GetUserIDFromContext(c)

	if err != nil {
		log.Warn().Msg("user ID not found in context")
		response.Unauthorized(c, "Unauthorized")
		return
	}

	req, err := requestutils.Bind[bookmarkDTO.ListBookmarksRequest](c)

	if err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	req.SetDefaults()

	bookmarks, pagination, err := h.service.List(c, userID, req)

	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Msg("failed to list bookmarks")

		response.InternalServerError(c)
		return
	}

	bookmarkDataList := make([]bookmarkDTO.BookmarkData, len(bookmarks))
	for i, bm := range bookmarks {
		bookmarkDataList[i] = bookmarkDTO.BookmarkData{
			ID:          bm.ID,
			Code:        bm.Code,
			Description: bm.Description,
			URL:         bm.URL,
			CreatedAt:   bm.CreatedAt,
			UpdatedAt:   bm.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response.Paginated(
		bookmarkDataList,
		pagination.Page,
		pagination.Limit,
		pagination.Total,
		"Bookmarks retrieved successfully!",
	))
}
