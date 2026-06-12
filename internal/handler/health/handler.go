package health

import (
	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-service/internal/service/health"
)

// Handler defines the interface for health check-related HTTP handlers, including the GetHealthCheck endpoint.
type Handler interface {
	GetHealthCheck(c *gin.Context)
}

type handler struct {
	service health.Service
}

// NewHandler creates a new instance of the health handler with the provided health service.
func NewHandler(service health.Service) Handler {
	return &handler{
		service: service,
	}
}
