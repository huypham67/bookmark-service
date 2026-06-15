package bookmark

type BookmarkCSVRecord struct {
	Description string `csv:"description" json:"description" validate:"required,min=2,max=500"`
	URL         string `csv:"url" json:"url" validate:"required,url"`
}

// BookmarkImportMessage is the message format sent to the import queue, containing a batch of bookmark records for a specific import job.
type BookmarkImportMessage struct {
	JobID   string              `json:"job_id"`  // Unique identifier for the import job
	UserID  string              `json:"user_id"` // ID of the user who initiated the import
	Records []BookmarkCSVRecord `json:"records"` // Batch of bookmark records to be imported
}
