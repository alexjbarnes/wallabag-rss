// Package wallabag provides a client for interacting with the Wallabag API for adding articles.
//
//nolint:cyclop // Package complexity is acceptable for API client functionality
package wallabag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	tokenURLPath = "/oauth/v2/token"
	entryURLPath = "/api/entries.json"
)

// Clienter defines the interface for Wallabag API interactions.
type Clienter interface {
	Authenticate(ctx context.Context) error
	AddEntry(ctx context.Context, urlToAdd string) (*Entry, error)
}

// Client represents the Wallabag API client.
type Client struct {
	expiresAt    time.Time
	httpClient   HTTPClient // Use an interface for http.Client
	baseURL      string
	clientID     string
	clientSecret string
	username     string
	password     string
	accessToken  string
}

// HTTPClient interface for mocking http.Client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewClient creates a new Wallabag API client.
func NewClient(baseURL, clientID, clientSecret, username, password string) *Client {
	return &Client{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		username:     username,
		password:     password,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// TokenResponse represents the response from the OAuth2 token endpoint.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// Entry represents a Wallabag entry.
type Entry struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	ID    int    `json:"id"`
}

// Authenticate performs OAuth2 authentication and sets the access token.
func (c *Client) Authenticate(ctx context.Context) error {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("username", c.username)
	data.Set("password", c.password)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+tokenURLPath, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't return since we're processing response
		}
	}()

	if resp.StatusCode != http.StatusOK {
		// Don't include response body in error to prevent information disclosure

		return fmt.Errorf("authentication failed with status %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// AddEntry adds a new entry to Wallabag.
func (c *Client) AddEntry(ctx context.Context, urlToAdd string) (*Entry, error) {
	if c.accessToken == "" || time.Now().After(c.expiresAt) {
		if err := c.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("failed to authenticate before adding entry: %w", err)
		}
	}

	entryData := map[string]string{"url": urlToAdd}
	jsonBody, err := json.Marshal(entryData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entry data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+entryURLPath, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create add entry request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send add entry request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't return since we're processing response
		}
	}()

	if resp.StatusCode != http.StatusOK {
		// Don't include response body in error to prevent information disclosure

		return nil, fmt.Errorf("failed to add entry with status %d", resp.StatusCode)
	}

	var entry Entry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, fmt.Errorf("failed to decode add entry response: %w", err)
	}

	return &entry, nil
}
