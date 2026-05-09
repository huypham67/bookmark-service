package model

// HealthCheckResponse represents the health check response payload.
type HealthCheckResponse struct {
	Message     string `json:"message"`
	ServiceName string `json:"service_name"`
	InstanceID  string `json:"instance_id"`
}
