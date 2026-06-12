package link

import (
	"context"
	"errors"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/redis/go-redis/v9"
)

// CheckExists checks whether the short code already exists.
func (r *repository) CheckExists(ctx context.Context, code string) (bool, error) {
	result, err := r.client.Exists(ctx, code).Result()
	if err != nil {
		return false, dbutils.ClassifyError(err)
	}
	return result > 0, nil
}

// GetLink retrieves original URL from Redis.
func (r *repository) GetLink(ctx context.Context, code string) (string, error) {
	url, err := r.client.Get(ctx, code).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", dbutils.ErrRecordNotFoundType
		}
		return "", err
	}
	return url, nil
}
