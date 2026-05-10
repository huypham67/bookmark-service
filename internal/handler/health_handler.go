package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/dto/response"
	"github.com/huypham67/bookmark-management/internal/service"
	"github.com/rs/zerolog/log"
)

// HealthCheck defines the contract for health check HTTP handlers.
type HealthCheck interface {
	GetHealthCheck(c *gin.Context)
}

type healthCheckHandler struct {
	healthCheckService service.HealthCheckService
}

// NewHealthCheckHandler creates a new health check handler with the given health check service.
func NewHealthCheckHandler(healthCheckService service.HealthCheckService) HealthCheck {
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
	var res response.HealthCheckResponse
	res = h.healthCheckService.GetStatus()
	if res.Message == "FAILED" {
		log.Error().
			Str("message", res.Message).
			Msg("500 - health check failed")

		c.JSON(http.StatusInternalServerError, res)
		return
	}
	c.JSON(http.StatusOK, res)
}
