package service

import (
	"github.com/huypham67/bookmark-management/internal/model"
)

const statusMessage = "OK"

// HealthCheck defines the contract for health check services.
type HealthCheck interface {
	GetStatus() model.HealthCheckResponse
}

type healthCheckService struct {
	serviceName string
	instanceID  string
}

// NewHealthCheckService creates a new health check service.
func NewHealthCheckService(serviceName string, instanceID string) HealthCheck {
	return &healthCheckService{
		serviceName: serviceName,
		instanceID:  instanceID,
	}
}

// GetStatus returns the current application health status.
func (s *healthCheckService) GetStatus() model.HealthCheckResponse {
	return model.HealthCheckResponse{
		Message:     statusMessage,
		ServiceName: s.serviceName,
		InstanceID:  s.instanceID,
	}
}
