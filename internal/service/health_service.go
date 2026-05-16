package service

import (
	"context"

	"github.com/huypham67/bookmark-service/internal/dto/response"
	"github.com/huypham67/bookmark-service/internal/repository"
	"github.com/rs/zerolog/log"
)

const statusMessage = "OK"
const failedStatusMessage = "FAILED"

// HealthCheckService defines the contract for health check services.
type HealthCheckService interface {
	GetStatus(ctx context.Context) response.HealthCheckResponse
}

type healthCheckService struct {
	serviceName string
	instanceID  string
	pinger      repository.Pinger
}

// NewHealthCheckService creates a new health check service.
func NewHealthCheckService(serviceName string, instanceID string, pinger repository.Pinger) HealthCheckService {
	return &healthCheckService{
		serviceName: serviceName,
		instanceID:  instanceID,
		pinger:      pinger,
	}
}

// GetStatus checks the health status of the application by pinging Redis and returns a HealthCheckResponse.
func (s *healthCheckService) GetStatus(ctx context.Context) response.HealthCheckResponse {
	if err := s.pinger.Ping(ctx); err != nil {
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
