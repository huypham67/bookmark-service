package bookmark

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/huypham67/bookmark-common/pkg/csvutil"
	"github.com/huypham67/bookmark-common/pkg/sliceutil"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/repository/queue"
	"github.com/rs/zerolog/log"
)

const (
	importBatchSize = 100
	importQueueName = "bookmark:import:jobs"
)

// Importer defines the interface for importing bookmarks from a CSV file.
// It is separate from Service because importing is a distinct use case with its
// own dependency (a queue publisher) and performance characteristics (batching,
// asynchronous job handling) — it never touches the bookmark repository.
//
//go:generate mockery --name=Importer --output=./mocks --outpkg=mocks --filename=mock_importer.go
type Importer interface {
	Import(ctx context.Context, userID string, r io.Reader) error
}

type importer struct {
	queuePublisher queue.Publisher
}

// NewImporter creates a new bookmark importer. The queue publisher is required:
// the importer's whole job is to hand work off to the queue.
func NewImporter(publisher queue.Publisher) Importer {
	return &importer{
		queuePublisher: publisher,
	}
}

// Import parses a bookmark CSV, splits the rows into batches, and enqueues one
// job message per batch for a worker to bulk-insert asynchronously.
//
// It does not write to the database itself: the API stays fast and returns
// once the work is safely on the queue. All batches share a single job ID so
// the worker can correlate (and de-duplicate) them.
func (i *importer) Import(ctx context.Context, userID string, r io.Reader) error {
	records, err := csvutil.Decode[bookmarkDTO.BookmarkCSVRecord](r)
	if err != nil {
		switch {
		case errors.Is(err, csvutil.ErrEmptyCSV):
			return ErrEmptyFile
		case errors.Is(err, csvutil.ErrInvalidFormat):
			log.Warn().
				Err(err).
				Str("user_id", userID).
				Msg("rejected invalid bookmark import csv")
			return ErrInvalidCSV
		default:
			log.Error().
				Err(err).
				Str("user_id", userID).
				Msg("failed to decode bookmark import csv")
			return ErrInternalServerError
		}
	}

	batches := sliceutil.SplitIntoBatches(records, importBatchSize)
	jobID := uuid.New().String()

	payloads := make([][]byte, 0, len(batches))
	for _, batch := range batches {
		job := bookmarkDTO.BookmarkImportMessage{
			JobID:   jobID,
			UserID:  userID,
			Records: batch,
		}

		payload, err := json.Marshal(job)
		if err != nil {
			log.Error().
				Err(err).
				Str("user_id", userID).
				Str("job_id", jobID).
				Msg("failed to marshal bookmark import job")
			return ErrInternalServerError
		}

		payloads = append(payloads, payload)
	}

	if err := i.queuePublisher.Enqueue(ctx, importQueueName, payloads...); err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Str("job_id", jobID).
			Msg("failed to enqueue bookmark import jobs")
		return ErrInternalServerError
	}

	log.Info().
		Str("job_id", jobID).
		Str("user_id", userID).
		Int("total_rows", len(records)).
		Int("total_batches", len(batches)).
		Msg("bookmark import jobs enqueued")

	return nil
}
