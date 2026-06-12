package link

import (
	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-service/internal/service/link"
)

// Handler defines the interface for link management-related HTTP handlers, including URL shortening and redirection.
type Handler interface {
	ShortenURL(c *gin.Context)
	RedirectToURL(c *gin.Context)
}

type handler struct {
	service link.Service
}

// NewHandler creates a new instance of the link handler with the provided link service.
func NewHandler(service link.Service) Handler {
	return &handler{
		service: service,
	}
}
