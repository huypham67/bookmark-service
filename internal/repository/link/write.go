package link

import (
	"context"
	"time"
)

// SaveLink stores a shortened URL in Redis with expiration time.
func (r *repository) SaveLink(ctx context.Context, code string, url string, exp int64) error {
	return r.client.Set(ctx, code, url, time.Duration(exp)*time.Second).Err()
}
