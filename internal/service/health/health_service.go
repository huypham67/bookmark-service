package health

import (
	"github.com/huypham67/bookmark-management/internal/dto/response"
)

const statusMessage = "OK"
const failedStatusMessage = "FAILED"

// HealthCheck defines the contract for health check services.
type HealthCheck interface {
	GetStatus() (response.HealthCheckResponse, error)
}

// Pinger defines the contract for health check ping operations.
type Pinger interface {
	Ping() error
}

type healthCheckService struct {
	serviceName string
	instanceID  string
	pinger      Pinger
}

// NewHealthCheckService creates a new health check service.
func NewHealthCheckService(serviceName string, instanceID string, pinger Pinger) HealthCheck {
	return &healthCheckService{
		serviceName: serviceName,
		instanceID:  instanceID,
		pinger:      pinger,
	}
}

// GetStatus returns the current application health status with service name, instance ID, and Redis connection status.
func (s *healthCheckService) GetStatus() (response.HealthCheckResponse, error) {
	if err := s.pinger.Ping(); err != nil {
		return response.HealthCheckResponse{
			Message:     failedStatusMessage,
			ServiceName: s.serviceName,
			InstanceID:  s.instanceID,
		}, err
	}

	return response.HealthCheckResponse{
		Message:     statusMessage,
		ServiceName: s.serviceName,
		InstanceID:  s.instanceID,
	}, nil
}
