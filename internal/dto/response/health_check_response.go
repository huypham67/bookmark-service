package response

// HealthCheckResponse represents the health check response payload with service status information.
type HealthCheckResponse struct {
	Message     string `json:"message"`
	ServiceName string `json:"service_name"`
	InstanceID  string `json:"instance_id"`
}
