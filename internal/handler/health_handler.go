package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/service"
)

// HealthCheck defines the contract for health check HTTP handlers.
type HealthCheck interface {
	GetHealthCheck(c *gin.Context)
}

type healthCheckHandler struct {
	healthCheckService service.HealthCheck
}

// NewHealthCheckHandler creates a new health check handler.
func NewHealthCheckHandler(healthCheckService service.HealthCheck) HealthCheck {
	return &healthCheckHandler{
		healthCheckService: healthCheckService,
	}
}

// GetHealthCheck handles the health check endpoint.
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
