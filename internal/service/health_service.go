package service

import (
	"github.com/huypham67/bookmark-management/internal/dto/response"
	"github.com/huypham67/bookmark-management/pkg/logger"
	"github.com/huypham67/bookmark-management/pkg/redis"
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

// GetStatus returns the current application health status with service name, instance ID, and Redis connection status.
func (s *healthCheckService) GetStatus() response.HealthCheckResponse {
	if err := s.pinger.Ping(); err != nil {
		logger.Get().Error().
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
