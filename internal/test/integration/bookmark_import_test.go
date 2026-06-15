package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// importQueueName mirrors the (unexported) queue name the service publishes to.
// The integration test treats it as the externally-observable Redis key.
const importQueueName = "bookmark:import:jobs"

// newImportRequest builds a multipart upload to the import endpoint.
func newImportRequest(t *testing.T, field, filename, content string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, filename)
	require.NoError(t, err)
	_, err = part.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/bookmark_service/v1/bookmarks/import",
		&body,
	)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestImportBookmarksEndpoint(t *testing.T) {
	t.Parallel()

	const importUserID = "user-uuid-1"

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name        string
		filename    string
		fixture     string
		withAuth    bool
		expected    expected
		verifyQueue func(t *testing.T, app *AuthenticatedTestApp)
	}{
		{
			name:     "should return 401 when the Authorization header is missing",
			filename: "bookmarks.csv",
			fixture:  fixtures.CSVBookmarksValid,
			withAuth: false,
			expected: expected{
				statusCode:   http.StatusUnauthorized,
				bodyContains: "missing authorization header",
			},
			verifyQueue: func(t *testing.T, app *AuthenticatedTestApp) {
				assert.False(t, app.MockRedis.Server.Exists(importQueueName), "no job should be enqueued")
			},
		},
		{
			name:     "should return 200 and enqueue the import job",
			filename: "bookmarks.csv",
			fixture:  fixtures.CSVBookmarksValid,
			withAuth: true,
			expected: expected{
				statusCode:   http.StatusOK,
				bodyContains: "Successfully sent bookmark imports to queue!",
			},
			verifyQueue: func(t *testing.T, app *AuthenticatedTestApp) {
				messages, err := app.MockRedis.Server.List(importQueueName)
				require.NoError(t, err)
				require.Len(t, messages, 1)

				var msg bookmarkDTO.BookmarkImportMessage
				require.NoError(t, json.Unmarshal([]byte(messages[0]), &msg))
				assert.Equal(t, importUserID, msg.UserID, "the JWT subject must own the import")
				assert.NotEmpty(t, msg.JobID)
				require.Len(t, msg.Records, 2)
			},
		},
		{
			name:     "should return 400 and enqueue nothing for an invalid csv",
			filename: "bookmarks.csv",
			fixture:  fixtures.CSVBookmarksInvalid,
			withAuth: true,
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid CSV file",
			},
			verifyQueue: func(t *testing.T, app *AuthenticatedTestApp) {
				assert.False(t, app.MockRedis.Server.Exists(importQueueName), "nothing should be enqueued on rejection")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := setupBookmarkTestApp(t)

			req := newImportRequest(t, "file", tc.filename, string(fixtures.ReadCSV(t, tc.fixture)))
			if tc.withAuth {
				token, err := app.TokenGenerator.GenerateToken(importUserID, "testuser1", "testuser1@gmail.com")
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			recorder := httptest.NewRecorder()
			app.Router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Contains(t, recorder.Body.String(), tc.expected.bodyContains)

			if tc.verifyQueue != nil {
				tc.verifyQueue(t, app)
			}
		})
	}
}
