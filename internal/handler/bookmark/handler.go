package bookmark

import (
	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
)

// Handler defines the interface for bookmark-related HTTP handlers.
type Handler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type handler struct {
	service bookmark.Service
}

// NewHandler creates a new instance of the bookmark handler with the provided bookmark service.
func NewHandler(service bookmark.Service) Handler {
	return &handler{
		service: service,
	}
}
