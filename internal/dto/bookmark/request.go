package bookmark

// CreateBookmarkRequest represents the request payload for creating a new bookmark.
type CreateBookmarkRequest struct {
	Description string `json:"description" binding:"required" validate:"required,min=2,max=500"`
	URL         string `json:"url" binding:"required" validate:"required,url"`
}

// UpdateBookmarkRequest handles multi-source binding: ID from URI, data from JSON.
//
// Binding sources:
// - ID: URI parameter (PUT /v1/bookmarks/:id)
// - Description, URL: JSON body
//
// Validation strategy:
// - Use binding tags ONLY for type conversion (omitted = optional at bind time)
// - Use validate tags for business logic validation (runs AFTER all sources bound)
type UpdateBookmarkRequest struct {
	ID          string `uri:"id" validate:"required"`                          // From URI, required
	Description string `json:"description" validate:"omitempty,min=2,max=500"` // From JSON, optional but if provided must be valid
	URL         string `json:"url" validate:"omitempty,url"`                   // From JSON, optional but if provided must be valid URL
}

// DeleteBookmarkRequest handles URI binding for DELETE endpoint.
type DeleteBookmarkRequest struct {
	ID string `uri:"id" validate:"required"`
}

// ListBookmarksRequest handles query parameters for listing bookmarks with pagination and sorting.
type ListBookmarksRequest struct {
	Page  int64  `form:"page" binding:"omitempty,min=1"`
	Limit int64  `form:"limit" binding:"omitempty,min=1,max=100"`
	Sort  string `form:"sort" binding:"omitempty,oneof=created_at updated_at code url"`
}

// SetDefaults assigns default values to pagination and sorting fields if they are not provided in the request.
func (r *ListBookmarksRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 10
	}
	if r.Sort == "" {
		r.Sort = "created_at"
	}
}
