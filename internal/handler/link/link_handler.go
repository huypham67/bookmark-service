package link

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/dto/request"
	"github.com/huypham67/bookmark-management/internal/dto/response"
	"github.com/huypham67/bookmark-management/internal/service/link"
)

// Link defines the contract for link HTTP handlers.
type Link interface {
	ShortenURL(c *gin.Context)
}

type linkHandler struct {
	linkService link.Link
}

// NewLinkHandler creates a new link handler with the given link service.
func NewLinkHandler(linkService link.Link) Link {
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
// @Success 200 {object} response.ShortenURLResponse
// @Failure 400 {object} gin.H "Invalid request body"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /links/shorten-url [post]
func (h *linkHandler) ShortenURL(c *gin.Context) {
	var req request.ShortenURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})

		return
	}

	code, err := h.linkService.ShortenURL(req)

	if err != nil {
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
