package bookmark

import (
	"time"

	"github.com/huypham67/bookmark-common/pkg/response"
)

type BookmarkData struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BookmarkResponse is a type alias for single bookmark response.
type BookmarkResponse = response.SuccessResponse[*BookmarkData]

// BookmarkListResponse is a type alias for paginated bookmark list response.
type BookmarkListResponse = response.SuccessResponse[[]BookmarkData]
