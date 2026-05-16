package service

import (
	"github.com/huypham67/bookmark-service/internal/dto/response"
	"github.com/huypham67/bookmark-service/pkg/redis"
	"github.com/rs/zerolog/log"
)

const statusMessage = "OK"
const failedStatusMessage = "FAILED"

// HealthCheckService defines the contract for health check services.
type HealthCheckService interface {
	GetStatus() response.HealthCheckResponse
}

type healthCheckService struct {
	serviceName string
	instanceID  string
	pinger      redis.Pinger
}

// NewHealthCheckService creates a new health check service.
func NewHealthCheckService(serviceName string, instanceID string, pinger redis.Pinger) HealthCheckService {
	return &healthCheckService{
		serviceName: serviceName,
		instanceID:  instanceID,
		pinger:      pinger,
	}
}

// GetStatus checks the health status of the application by pinging Redis and returns a HealthCheckResponse.
func (s *healthCheckService) GetStatus() response.HealthCheckResponse {
	if err := s.pinger.Ping(); err != nil {
		log.Error().
			Err(err).
			Str("service", s.serviceName).
			Msg("Redis connection failed - health check")

		return response.HealthCheckResponse{
			Message:     failedStatusMessage,
			ServiceName: s.serviceName,
			InstanceID:  s.instanceID,
		}
	}

	return response.HealthCheckResponse{
		Message:     statusMessage,
		ServiceName: s.serviceName,
		InstanceID:  s.instanceID,
	}
}
