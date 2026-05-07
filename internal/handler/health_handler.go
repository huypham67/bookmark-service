package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/service"
)

// HealthCheck defines the interface for the health check handler.
type HealthCheck interface {
	GetHealthCheck(c *gin.Context)
}

type healthCheckHandler struct {
	healthCheckService service.HealthCheck
}

// NewHealthCheckHandler creates a new instance of the health check handler with the provided health check service.
func NewHealthCheckHandler(healthCheckService service.HealthCheck) HealthCheck {
	return &healthCheckHandler{
		healthCheckService: healthCheckService,
	}
}

// GetHealthCheck godoc
//
// @Summary Health Check
// @Description Check application health status
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} model.HealthCheckResponse
// @Router /health-check [get]
func (h *healthCheckHandler) GetHealthCheck(c *gin.Context) {
	response := h.healthCheckService.GetStatus()
	c.JSON(http.StatusOK, response)
}
