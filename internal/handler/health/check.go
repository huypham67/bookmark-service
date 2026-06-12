package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	healthDTO "github.com/huypham67/bookmark-service/internal/dto/health"
	"github.com/rs/zerolog/log"
)

var _ healthDTO.HealthCheckResponse

// GetHealthCheck handles the health check endpoint.
//
// @Summary Health Check
// @Description Check application health status and Redis connection
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} healthDTO.HealthCheckResponse
// @Failure 500 {object} healthDTO.HealthCheckResponse
// @Router /health-check [get]
func (h *handler) GetHealthCheck(c *gin.Context) {
	res := h.service.GetStatus(c)
	if res.Message == "FAILED" {
		log.Error().
			Str("message", res.Message).
			Msg("500 - health check failed")

		c.JSON(http.StatusInternalServerError, res)
		return
	}
	c.JSON(http.StatusOK, res)
}
