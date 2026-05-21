package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	// Response DTO is imported for Swagger documentation purposes, even though it's not directly used in the code.
	_ "github.com/huypham67/bookmark-service/internal/dto/response"

	"github.com/huypham67/bookmark-service/internal/service"
	"github.com/rs/zerolog/log"
)

// HealthCheck defines the contract for health check HTTP handlers.
type HealthCheck interface {
	GetHealthCheck(c *gin.Context)
}

type healthCheckHandler struct {
	healthCheckService service.HealthCheck
}

// NewHealthCheckHandler creates a new health check handler with the given health check service.
func NewHealthCheckHandler(healthCheckService service.HealthCheck) HealthCheck {
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
	res := h.healthCheckService.GetStatus(c)
	if res.Message == "FAILED" {
		log.Error().
			Str("message", res.Message).
			Msg("500 - health check failed")

		c.JSON(http.StatusInternalServerError, res)
		return
	}
	c.JSON(http.StatusOK, res)
}
