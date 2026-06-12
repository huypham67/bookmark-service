package link

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/requestutils"
	"github.com/huypham67/bookmark-common/pkg/response"
	linkDTO "github.com/huypham67/bookmark-service/internal/dto/link"
	"github.com/rs/zerolog/log"
)

// ShortenURL handles the URL shortening endpoint.
//
// @Summary Shorten URL
// @Description Create a shortened URL code and save it to Redis
// @Tags links
// @Accept json
// @Produce json
// @Param request body linkDTO.ShortenURLRequest true "URL to shorten"
// @Success 200 {object} linkDTO.ShortenURLResponse "Shorten URL generated successfully"
// @Failure 400 {object} gin.H "Invalid request body"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/links/shorten [post]
func (h *handler) ShortenURL(c *gin.Context) {
	req, err := requestutils.Bind[linkDTO.ShortenURLRequest](c)

	if err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	code, err := h.service.ShortenURL(c, *req)

	if err != nil {
		log.Error().
			Err(err).
			Str("url", req.Url).
			Int64("exp", req.Exp).
			Msg("failed to shorten URL")

		response.InternalServerError(c)

		return
	}

	c.JSON(http.StatusOK, linkDTO.ShortenURLResponse{
		Code:    code,
		Message: "Shorten URL generated successfully",
	})
}
