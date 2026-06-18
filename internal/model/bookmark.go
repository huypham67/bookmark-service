package model

// Bookmark represents a user's bookmark with fields for description, URL, unique code, and associated user ID.
type Bookmark struct {
	BaseModel
	Description string `json:"description" gorm:"type:text"`
	URL         string `json:"url" gorm:"type:text"`
	Code        string `json:"code" gorm:"type:varchar(255)"`
	CodeInt     int    `json:"code_int" gorm:"type:serial;uniqueIndex;autoIncrement"`
	UserID      string `json:"user_id" gorm:"type:uuid;not null;index"`
}
