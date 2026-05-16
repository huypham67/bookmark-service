package response

// ShortenURLResponse represents the response payload for URL shortening endpoint with the generated short code.
type ShortenURLResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
