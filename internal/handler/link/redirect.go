package link

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/huypham67/bookmark-common/pkg/requestutils"
	"github.com/huypham67/bookmark-common/pkg/response"
	"github.com/huypham67/bookmark-common/pkg/shortcode"
	linkDTO "github.com/huypham67/bookmark-service/internal/dto/link"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog/log"
)

const (
	metricLinkRedirect = "Custom/Link/Redirect"
	metricLinkNotFound = "Custom/Link/RedirectNotFound"
)

// linkSource maps a code's routing store to a stable label for the redirect metric.
func linkSource(code string) string {
	switch shortcode.Classify(code) {
	case shortcode.StoreRedis:
		return "redis"
	case shortcode.StoreSQL:
		return "pg"
	default:
		return "unknown"
	}
}

// RedirectToURL handles the redirect endpoint.
//
// @Summary Redirect to Original URL
// @Description Redirect user to the original URL for a given code. The code's routing prefix selects the backing store: shortened links resolve from Redis, bookmarks resolve from the database.
// @Tags links
// @Accept json
// @Produce json
// @Param code path string true "Short code (prefixed: a-h for links, i-z for bookmarks)"
// @Success 302 "Redirect successful"
// @Failure 404 {object} gin.H "Short link not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/links/redirect/{code} [get]
func (h *handler) RedirectToURL(c *gin.Context) {
	txn := newrelic.FromContext(c)
	defer txn.StartSegment("handler.link.RedirectToURL").End()

	req, err := requestutils.Bind[linkDTO.RedirectRequest](c)

	if err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	code := req.Code

	url, err := h.service.GetOriginalURL(c, code)

	if err != nil {
		if errors.Is(err, dbutils.ErrRecordNotFoundType) {
			txn.Application().RecordCustomMetric(metricLinkNotFound, 1)
			response.NotFound(c, "Short link not found")
			return
		}

		log.Error().
			Err(err).
			Str("code", code).
			Msg("failed to retrieve original URL")

		response.InternalServerError(c)
		return
	}

	txn.Application().RecordCustomMetric(metricLinkRedirect, 1)
	txn.AddAttribute("link.source", linkSource(code))

	c.Redirect(http.StatusFound, url)
}
