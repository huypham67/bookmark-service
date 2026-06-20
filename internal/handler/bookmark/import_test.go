package bookmark

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/huypham67/bookmark-service/internal/service/bookmark/mocks"
	"github.com/huypham67/bookmark-service/internal/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// multipartFileRequest builds a POST with a single file form field.
func multipartFileRequest(t *testing.T, field, filename, content string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, filename)
	require.NoError(t, err)
	_, err = part.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/v1/bookmarks/import", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func setUser(ctx *gin.Context) {
	ctx.Set("claims", &jwt.CustomClaims{UserID: "user-id-123"})
}

func TestHandler_Import(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name         string
		buildRequest func(*testing.T) *http.Request
		setupClaims  func(*gin.Context)
		setupMock    func(context.Context, *mocks.Importer)
		expected     expected
	}{
		{
			name: "should return 200 when import is accepted",
			buildRequest: func(t *testing.T) *http.Request {
				return multipartFileRequest(t, "file", "bookmarks.csv", string(fixtures.ReadCSV(t, fixtures.CSVBookmarksValid)))
			},
			setupClaims: setUser,
			setupMock: func(ctx context.Context, mockSvc *mocks.Importer) {
				mockSvc.On("Import", ctx, "user-id-123", mock.MatchedBy(func(r io.Reader) bool { return r != nil })).
					Return(nil).Once()
			},
			expected: expected{
				statusCode:   http.StatusOK,
				bodyContains: "Successfully sent bookmark imports to queue!",
			},
		},
		{
			name: "should return 401 when user ID is not in context",
			buildRequest: func(t *testing.T) *http.Request {
				return multipartFileRequest(t, "file", "bookmarks.csv", string(fixtures.ReadCSV(t, fixtures.CSVBookmarksValid)))
			},
			setupClaims: func(*gin.Context) {},
			setupMock:   func(ctx context.Context, mockSvc *mocks.Importer) {},
			expected: expected{
				statusCode:   http.StatusUnauthorized,
				bodyContains: "Unauthorized",
			},
		},
		{
			name: "should return 400 when file field is missing",
			buildRequest: func(t *testing.T) *http.Request {
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)
				require.NoError(t, writer.WriteField("notfile", "x"))
				require.NoError(t, writer.Close())
				req := httptest.NewRequest(http.MethodPost, "/v1/bookmarks/import", &body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			setupClaims: setUser,
			setupMock:   func(ctx context.Context, mockSvc *mocks.Importer) {},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "file is required",
			},
		},
		{
			name: "should return 400 when file is not a .csv",
			buildRequest: func(t *testing.T) *http.Request {
				return multipartFileRequest(t, "file", "bookmarks.txt", string(fixtures.ReadCSV(t, fixtures.CSVBookmarksValid)))
			},
			setupClaims: setUser,
			setupMock:   func(ctx context.Context, mockSvc *mocks.Importer) {},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "only .csv files are allowed",
			},
		},
		{
			name: "should return 413 when file exceeds the size limit",
			buildRequest: func(t *testing.T) *http.Request {
				// Header + one oversized row to push the body past maxImportFileSize.
				big := "description,url\n" + strings.Repeat("a", maxImportFileSize+1)
				return multipartFileRequest(t, "file", "big.csv", big)
			},
			setupClaims: setUser,
			setupMock:   func(ctx context.Context, mockSvc *mocks.Importer) {},
			expected: expected{
				statusCode:   http.StatusRequestEntityTooLarge,
				bodyContains: "File too large",
			},
		},
		{
			name: "should return 400 when service reports an empty file",
			buildRequest: func(t *testing.T) *http.Request {
				return multipartFileRequest(t, "file", "bookmarks.csv", string(fixtures.ReadCSV(t, fixtures.CSVBookmarksEmpty)))
			},
			setupClaims: setUser,
			setupMock: func(ctx context.Context, mockSvc *mocks.Importer) {
				mockSvc.On("Import", ctx, "user-id-123", mock.MatchedBy(func(r io.Reader) bool { return r != nil })).
					Return(bookmark.ErrEmptyFile).Once()
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "CSV file is empty",
			},
		},
		{
			name: "should return 400 when service reports invalid csv",
			buildRequest: func(t *testing.T) *http.Request {
				return multipartFileRequest(t, "file", "bookmarks.csv", string(fixtures.ReadCSV(t, fixtures.CSVBookmarksValid)))
			},
			setupClaims: setUser,
			setupMock: func(ctx context.Context, mockSvc *mocks.Importer) {
				mockSvc.On("Import", ctx, "user-id-123", mock.MatchedBy(func(r io.Reader) bool { return r != nil })).
					Return(bookmark.ErrInvalidCSV).Once()
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid CSV file",
			},
		},
		{
			name: "should return 500 when service returns an unexpected error",
			buildRequest: func(t *testing.T) *http.Request {
				return multipartFileRequest(t, "file", "bookmarks.csv", string(fixtures.ReadCSV(t, fixtures.CSVBookmarksValid)))
			},
			setupClaims: setUser,
			setupMock: func(ctx context.Context, mockSvc *mocks.Importer) {
				mockSvc.On("Import", ctx, "user-id-123", mock.MatchedBy(func(r io.Reader) bool { return r != nil })).
					Return(bookmark.ErrInternalServerError).Once()
			},
			expected: expected{
				statusCode:   http.StatusInternalServerError,
				bodyContains: "Internal Server Error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			mockSvc := new(mocks.Importer)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = tc.buildRequest(t)

			tc.setupClaims(ctx)
			tc.setupMock(ctx, mockSvc)

			handler := NewHandler(nil, mockSvc)
			handler.Import(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Contains(t, recorder.Body.String(), tc.expected.bodyContains)

			mockSvc.AssertExpectations(t)
		})
	}
}
