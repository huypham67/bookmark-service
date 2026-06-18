package bookmark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	pkgRedis "github.com/huypham67/bookmark-common/pkg/redis"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/repository/queue"
	"github.com/huypham67/bookmark-service/internal/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newImportService builds an importer with a real queue publisher backed by
// miniredis, so tests can assert the actual messages that land on the queue.
func newImportService(t *testing.T) (Importer, *pkgRedis.Mock) {
	t.Helper()

	mockRedis := pkgRedis.NewMock(t)
	svc := NewImporter(queue.NewRedisPublisher(mockRedis.Client))

	return svc, mockRedis
}

// buildImportCSV generates a valid import CSV with n data rows. Used only for
// the batching case, where the row count (not the content) is what matters.
func buildImportCSV(n int) []byte {
	var b strings.Builder
	b.WriteString("description,url\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "Item %d,https://example.com/%d\n", i, i)
	}
	return []byte(b.String())
}

func TestService_Import(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		buildInput func(*testing.T) []byte
		setup      func(*pkgRedis.Mock)
		verify     func(*testing.T, *pkgRedis.Mock, error)
	}{
		{
			name: "enqueues a single batch for a small file",
			buildInput: func(t *testing.T) []byte {
				return fixtures.ReadCSV(t, fixtures.CSVBookmarksValid)
			},
			verify: func(t *testing.T, mockRedis *pkgRedis.Mock, err error) {
				require.NoError(t, err)

				messages, lerr := mockRedis.Server.List(importQueueName)
				require.NoError(t, lerr)
				require.Len(t, messages, 1, "2 rows fit in one batch")

				var msg bookmarkDTO.BookmarkImportMessage
				require.NoError(t, json.Unmarshal([]byte(messages[0]), &msg))
				assert.Equal(t, "user-1", msg.UserID)
				assert.NotEmpty(t, msg.JobID)
				require.Len(t, msg.Records, 2)
				assert.Equal(t, "Search engine", msg.Records[0].Description)
				assert.Equal(t, "https://google.com", msg.Records[0].URL)
			},
		},
		{
			name: "splits a large file into batches sharing one job id",
			buildInput: func(*testing.T) []byte {
				return buildImportCSV(importBatchSize*2 + 5) // -> 3 messages
			},
			verify: func(t *testing.T, mockRedis *pkgRedis.Mock, err error) {
				require.NoError(t, err)

				messages, lerr := mockRedis.Server.List(importQueueName)
				require.NoError(t, lerr)
				require.Len(t, messages, 3)

				var jobID string
				rowsSeen := 0
				for i, raw := range messages {
					var msg bookmarkDTO.BookmarkImportMessage
					require.NoError(t, json.Unmarshal([]byte(raw), &msg))
					if i == 0 {
						jobID = msg.JobID
					}
					assert.Equal(t, jobID, msg.JobID, "all batches of one import share a job id")
					rowsSeen += len(msg.Records)
				}
				assert.Equal(t, importBatchSize*2+5, rowsSeen, "every row must be accounted for")
			},
		},
		{
			name: "rejects an invalid row without enqueuing",
			buildInput: func(t *testing.T) []byte {
				return fixtures.ReadCSV(t, fixtures.CSVBookmarksInvalid)
			},
			verify: func(t *testing.T, mockRedis *pkgRedis.Mock, err error) {
				assert.ErrorIs(t, err, ErrInvalidCSV)
				assert.False(t, mockRedis.Server.Exists(importQueueName), "nothing enqueued on validation failure")
			},
		},
		{
			name: "rejects an empty file without enqueuing",
			buildInput: func(t *testing.T) []byte {
				return fixtures.ReadCSV(t, fixtures.CSVBookmarksEmpty)
			},
			verify: func(t *testing.T, mockRedis *pkgRedis.Mock, err error) {
				assert.ErrorIs(t, err, ErrEmptyFile)
				assert.False(t, mockRedis.Server.Exists(importQueueName))
			},
		},
		{
			name: "returns internal error when the queue is unavailable",
			buildInput: func(t *testing.T) []byte {
				return fixtures.ReadCSV(t, fixtures.CSVBookmarksValid)
			},
			setup: func(mockRedis *pkgRedis.Mock) {
				mockRedis.Close() // make every Redis call fail
			},
			verify: func(t *testing.T, _ *pkgRedis.Mock, err error) {
				assert.ErrorIs(t, err, ErrInternalServerError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, mockRedis := newImportService(t)
			if tc.setup != nil {
				tc.setup(mockRedis)
			}

			err := svc.Import(context.Background(), "user-1", bytes.NewReader(tc.buildInput(t)))

			tc.verify(t, mockRedis, err)
		})
	}
}
