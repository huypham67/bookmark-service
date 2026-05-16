package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-service/internal/dto/request"
	"github.com/huypham67/bookmark-service/internal/dto/response"
	"github.com/huypham67/bookmark-service/internal/service"
	"github.com/rs/zerolog/log"
)

// Link defines the contract for link HTTP handlers.
type Link interface {
	ShortenURL(c *gin.Context)
	RedirectToURL(c *gin.Context)
}

type linkHandler struct {
	linkService service.LinkService
}

// NewLinkHandler creates a new link handler with the given link service.
func NewLinkHandler(linkService service.LinkService) Link {
	return &linkHandler{
		linkService: linkService,
	}
}

// ShortenURL handles the URL shortening endpoint.
//
// @Summary Shorten URL
// @Description Create a shortened URL code and save it to Redis
// @Tags links
// @Accept json
// @Produce json
// @Param request body request.ShortenURLRequest true "URL to shorten"
// @Success 200 {object} response.ShortenURLResponse "Shorten URL generated successfully"
// @Failure 400 {object} gin.H "Invalid request body"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/links/shorten [post]
func (h *linkHandler) ShortenURL(c *gin.Context) {
	var req request.ShortenURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	code, err := h.linkService.ShortenURL(c, req)

	if err != nil {
		log.Error().
			Err(err).
			Str("url", req.Url).
			Int64("exp", req.Exp).
			Msg("500 - failed to shorten URL")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})

		return
	}

	c.JSON(http.StatusOK, response.ShortenURLResponse{
		Code:    code,
		Message: "Shorten URL generated successfully",
	})
}

// RedirectToURL handles the redirect endpoint.
//
// @Summary Redirect to Original URL
// @Description Redirect user to the original URL based on the shortened code
// @Tags links
// @Accept json
// @Produce json
// @Param code path string true "Shortened code"
// @Success 301 "Redirect successful"
// @Failure 404 {object} gin.H "Short link not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /v1/links/redirect/{code} [get]
func (h *linkHandler) RedirectToURL(c *gin.Context) {
	code := c.Param("code")

	url, err := h.linkService.GetOriginalURL(c, code)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Short link not found",
		})
		return
	}

	c.Redirect(http.StatusFound, url)
}
