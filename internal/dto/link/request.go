package link

// ShortenURLRequest represents the request payload for URL shortening endpoint.
type ShortenURLRequest struct {
	Url string `json:"url" binding:"required,url,max=2048"`
	Exp int64  `json:"exp" binding:"gte=0,lte=86400"`
}

// RedirectRequest handles URI binding for the redirect endpoint.
type RedirectRequest struct {
	Code string `uri:"code" validate:"required"`
}
