package request

// ShortenURLRequest represents the request payload for URL shortening endpoint.
type ShortenURLRequest struct {
	Url string `json:"url"`
	Exp int64  `json:"exp"`
}
