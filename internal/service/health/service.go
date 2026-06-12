package health

import (
	"context"

	healthDTO "github.com/huypham67/bookmark-service/internal/dto/health"
	"github.com/huypham67/bookmark-service/internal/repository/ping"
)

const (
	statusMessage       = "OK"
	failedStatusMessage = "FAILED"
)

// Service defines the contract for health check operations.
//
//go:generate mockery --name=Service --output=./mocks --outpkg=mocks --filename=mock_service.go
type Service interface {
	GetStatus(ctx context.Context) healthDTO.HealthCheckResponse
}

type service struct {
	serviceName string
	instanceID  string
	pinger      ping.Pinger
}

// NewService creates a new health check service with the provided configuration and pinger.
func NewService(serviceName string, instanceID string, pinger ping.Pinger) Service {
	return &service{
		serviceName: serviceName,
		instanceID:  instanceID,
		pinger:      pinger,
	}
}
