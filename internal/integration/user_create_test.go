package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huypham67/bookmark-service/internal/dto/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUserEndpoint(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name        string
		requestBody string
		expected    expected
	}{
		{
			name: "should return 201 when user creation is successful",
			requestBody: `{
				"display_name": "New User",
				"username": "newuser",
				"email": "newuser@example.com",
				"password": "password123"
			}`,
			expected: expected{
				statusCode:   http.StatusCreated,
				bodyContains: "Register an user successfully!",
			},
		},
		{
			name: "should return 400 when request body is invalid JSON",
			requestBody: `{invalid json}`,
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 400 when validation fails - missing display_name",
			requestBody: `{
				"username": "newuser",
				"email": "newuser@example.com",
				"password": "password123"
			}`,
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 400 when validation fails - missing email",
			requestBody: `{
				"display_name": "New User",
				"username": "newuser",
				"password": "password123"
			}`,
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 400 when validation fails - missing username",
			requestBody: `{
				"display_name": "New User",
				"email": "newuser@example.com",
				"password": "password123"
			}`,
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 400 when validation fails - missing password",
			requestBody: `{
				"display_name": "New User",
				"username": "newuser",
				"email": "newuser@example.com"
			}`,
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 409 when email already exists",
			requestBody: `{
				"display_name": "Duplicate User",
				"username": "differentusername",
				"email": "testuser1@gmail.com",
				"password": "password123"
			}`,
			expected: expected{
				statusCode:   http.StatusConflict,
				bodyContains: "User already exists",
			},
		},
		{
			name: "should return 409 when username already exists",
			requestBody: `{
				"display_name": "Duplicate User",
				"username": "testuser1",
				"email": "differentemail@example.com",
				"password": "password123"
			}`,
			expected: expected{
				statusCode:   http.StatusConflict,
				bodyContains: "User already exists",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := setupUserTestApp(t)

			httpRequest := httptest.NewRequest(http.MethodPost, "/api/bookmark_service/v1/users/register", bytes.NewBufferString(tc.requestBody))

			httpRequest.Header.Set("Content-Type", "application/json")
			httpRecorder := httptest.NewRecorder()
			app.Router.ServeHTTP(httpRecorder, httpRequest)

			assert.Equal(t, tc.expected.statusCode, httpRecorder.Code)
			assert.Contains(t, httpRecorder.Body.String(), tc.expected.bodyContains)

			if tc.expected.statusCode == http.StatusCreated {
				var resp response.RegisterUserResponse
				err := json.Unmarshal(httpRecorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp.Data.ID)
				assert.Equal(t, "New User", resp.Data.DisplayName)
				assert.Equal(t, "newuser", resp.Data.Username)
				assert.Equal(t, "newuser@example.com", resp.Data.Email)
			}
		})
	}
}
