package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/service/health"
)

// HealthCheck defines the contract for health check HTTP handlers.
type HealthCheck interface {
	GetHealthCheck(c *gin.Context)
}

type healthCheckHandler struct {
	healthCheckService health.HealthCheck
}

// NewHealthCheckHandler creates a new health check handler with the given health check service.
func NewHealthCheckHandler(healthCheckService health.HealthCheck) HealthCheck {
	return &healthCheckHandler{
		healthCheckService: healthCheckService,
	}
}

// GetHealthCheck handles the health check endpoint.
//
// @Summary Health Check
// @Description Check application health status and Redis connection
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.HealthCheckResponse
// @Failure 500 {object} response.HealthCheckResponse
// @Router /health-check [get]
func (h *healthCheckHandler) GetHealthCheck(c *gin.Context) {
	response, err := h.healthCheckService.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	c.JSON(http.StatusOK, response)
}
