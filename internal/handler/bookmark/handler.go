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
	Import(c *gin.Context)
}

type handler struct {
	service  bookmark.Service
	importer bookmark.Importer
}

// NewHandler creates a new bookmark handler. The CRUD endpoints use the
// (cache-decorated) service; Import uses the importer, which bypasses the cache.
func NewHandler(service bookmark.Service, importer bookmark.Importer) Handler {
	return &handler{
		service:  service,
		importer: importer,
	}
}
