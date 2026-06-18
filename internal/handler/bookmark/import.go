package bookmark

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	"github.com/huypham67/bookmark-common/pkg/requestutils"
	"github.com/huypham67/bookmark-common/pkg/response"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/rs/zerolog/log"
)

const (
	// importFileFormKey is the multipart form field that carries the CSV file.
	importFileFormKey = "file"

	// maxImportFileSize caps the upload. Rows are held in memory for validation
	// and batching before anything is enqueued, so the size limit is the main
	// guard against a huge upload exhausting memory.
	maxImportFileSize = 5 << 20 // 5 MiB
)

// Import handles the bookmark CSV import endpoint.
//
// @Summary Import Bookmarks
// @Description Upload a CSV file (columns: description,url) to be imported asynchronously
// @Tags bookmarks
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param file formData file true "CSV file with description,url columns"
// @Success 200 {object} response.MessageResponse "Import accepted"
// @Failure 400 {object} gin.H "Invalid file or content"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 413 {object} gin.H "File too large"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/bookmarks/import [post]
func (h *handler) Import(c *gin.Context) {
	userID, err := jwt.GetUserIDFromContext(c)
	if err != nil {
		log.Warn().Msg("user ID not found in context")
		response.Unauthorized(c, "Unauthorized")
		return
	}

	file, _, err := requestutils.FormFile(c, requestutils.FileConstraint{
		FormKey:      importFileFormKey,
		MaxSizeBytes: maxImportFileSize,
		AllowedExts:  []string{".csv"},
	})
	if err != nil {
		switch {
		case errors.Is(err, requestutils.ErrFileTooLarge):
			response.Error(c, http.StatusRequestEntityTooLarge, "File too large")
		case errors.Is(err, requestutils.ErrFileTypeDenied):
			response.BadRequest(c, "only .csv files are allowed")
		case errors.Is(err, requestutils.ErrFileRequired):
			response.BadRequest(c, "file is required")
		default:
			log.Error().Err(err).Str("user_id", userID).Msg("failed to read uploaded import file")
			response.InternalServerError(c)
		}
		return
	}
	defer file.Close()

	if err := h.importer.Import(c, userID, file); err != nil {
		switch {
		case errors.Is(err, bookmark.ErrEmptyFile):
			response.BadRequest(c, "CSV file is empty")
		case errors.Is(err, bookmark.ErrInvalidCSV):
			response.BadRequest(c, "Invalid CSV file")
		default:
			log.Error().Err(err).Str("user_id", userID).Msg("failed to import bookmarks")
			response.InternalServerError(c)
		}
		return
	}

	c.JSON(http.StatusOK, response.Message("Successfully sent bookmark imports to queue!"))
}
