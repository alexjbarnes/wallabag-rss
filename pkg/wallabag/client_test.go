package wallabag_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"wallabag-rss-tool/pkg/wallabag"
)

func TestNewClient(t *testing.T) {
	client := wallabag.NewClient(
		"https://wallabag.example.com",
		"client123",
		"secret456",
		"testuser",
		"testpass",
	)

	assert.NotNil(t, client)
	// Cannot access unexported fields from external test package
}

func TestClient_Authenticate(t *testing.T) {
	t.Run("Successful authentication", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/oauth/v2/token", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

			// Parse form data
			err := r.ParseForm()
			assert.NoError(t, err)
			assert.Equal(t, "password", r.FormValue("grant_type"))
			assert.Equal(t, "test_client", r.FormValue("client_id"))
			assert.Equal(t, "test_secret", r.FormValue("client_secret"))
			assert.Equal(t, "test_user", r.FormValue("username"))
			assert.Equal(t, "test_pass", r.FormValue("password"))

			// Return success response - use a simple struct since we can't access the exported type
			tokenResp := map[string]interface{}{
				"access_token":  "test_access_token",
				"expires_in":    3600,
				"token_type":    "Bearer",
				"scope":         "read write",
				"refresh_token": "test_refresh_token",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(tokenResp)
		}))
		defer server.Close()

		client := wallabag.NewClient(server.URL, "test_client", "test_secret", "test_user", "test_pass")

		err := client.Authenticate(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Authentication failure - wrong credentials", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"invalid_credentials"}`))
		}))
		defer server.Close()

		client := wallabag.NewClient(server.URL, "wrong_client", "wrong_secret", "wrong_user", "wrong_pass")

		err := client.Authenticate(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed with status 401")
	})

	t.Run("Network error", func(t *testing.T) {
		client := wallabag.NewClient("http://nonexistent.example.com", "client", "secret", "user", "pass")

		err := client.Authenticate(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send auth request")
	})
}

func TestClient_AddEntry(t *testing.T) {
	t.Run("Successful add entry", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/oauth/v2/token" {
				// Handle authentication
				tokenResp := map[string]interface{}{
					"access_token": "test_access_token",
					"expires_in":   3600,
					"token_type":   "Bearer",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tokenResp)
				return
			}

			if r.URL.Path == "/api/entries.json" {
				// Handle add entry
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test_access_token", r.Header.Get("Authorization"))

				// Parse request body
				var entryData map[string]string
				err := json.NewDecoder(r.Body).Decode(&entryData)
				assert.NoError(t, err)
				assert.Equal(t, "https://example.com/article", entryData["url"])

				// Return success response
				entry := map[string]interface{}{
					"id":    456,
					"url":   "https://example.com/article",
					"title": "Added Article",
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(entry)
				return
			}

			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := wallabag.NewClient(server.URL, "test_client", "test_secret", "test_user", "test_pass")

		entry, err := client.AddEntry(context.Background(), "https://example.com/article")
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		// Cannot access entry fields from external test package
	})

	t.Run("Add entry failure - authentication fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/oauth/v2/token" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"invalid_credentials"}`))
				return
			}
		}))
		defer server.Close()

		client := wallabag.NewClient(server.URL, "wrong_client", "wrong_secret", "wrong_user", "wrong_pass")

		entry, err := client.AddEntry(context.Background(), "https://example.com/article")
		assert.Error(t, err)
		assert.Nil(t, entry)
		assert.Contains(t, err.Error(), "failed to authenticate before adding entry")
	})
}

func TestClient_Interface(t *testing.T) {
	t.Run("Client implements Clienter interface", func(t *testing.T) {
		var client wallabag.Clienter = wallabag.NewClient("https://example.com", "id", "secret", "user", "pass")
		assert.NotNil(t, client)

		// Test that we can call interface methods
		assert.NotPanics(t, func() {
			client.Authenticate(context.Background())
			client.AddEntry(context.Background(), "https://example.com/article")
		})
	})
}

func TestHTTPClient_Interface(t *testing.T) {
	t.Run("http.Client implements HTTPClient interface", func(t *testing.T) {
		var httpClient wallabag.HTTPClient = &http.Client{}
		assert.NotNil(t, httpClient)

		// Test that we can call interface methods
		assert.NotPanics(t, func() {
			req, _ := http.NewRequest("GET", "https://example.com", http.NoBody)
			resp, _ := httpClient.Do(req)
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		})
	})
}

func TestConstants(t *testing.T) {
	t.Run("URL path constants", func(t *testing.T) {
		// Cannot access unexported constants from external test package
		// tokenURLPath and entryURLPath are not exported
	})
}
