package service

import (
	"github.com/huypham67/bookmark-management/internal/model"
)

const statusMessage = "OK"

// HealthCheck defines the interface for the health check service.
type HealthCheck interface {
	GetStatus() model.HealthCheckResponse
}

type healthCheckService struct {
	serviceName string
	instanceID  string
}

// NewHealthCheckService creates a new instance of the health check service with the provided service name and instance ID.
func NewHealthCheckService(serviceName string, instanceID string) HealthCheck {
	return &healthCheckService{
		serviceName: serviceName,
		instanceID:  instanceID,
	}
}

// GetStatus returns the health check status, including a message, service name, and instance ID.
func (s *healthCheckService) GetStatus() model.HealthCheckResponse {
	return model.HealthCheckResponse{
		Message:     statusMessage,
		ServiceName: s.serviceName,
		InstanceID:  s.instanceID,
	}
}
