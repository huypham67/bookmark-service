package request

// ShortenURLRequest represents the request payload for URL shortening endpoint.
type ShortenURLRequest struct {
	Url string `json:"url" validate:"required,url,max=2048"`
	Exp int64  `json:"exp" validate:"gte=0,lte=86400"`
}
