// Package server provides HTTP server functionality with CSRF protection and security middleware.
package server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const (
	httpMethodPUT    = "PUT"
	httpMethodPOST   = "POST"
	httpMethodDELETE = "DELETE"
)

type CSRFToken struct {
	ExpiresAt time.Time
	Value     string
}

type CSRFManager struct {
	tokens map[string]CSRFToken
	mutex  sync.RWMutex
}

func NewCSRFManager() *CSRFManager {
	manager := &CSRFManager{
		tokens: make(map[string]CSRFToken),
	}
	// Clean up expired tokens every hour
	go manager.cleanupExpiredTokens()

	return manager
}

func (c *CSRFManager) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(bytes)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.tokens[token] = CSRFToken{
		Value:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour expiry
	}

	return token, nil
}

func (c *CSRFManager) ValidateToken(token string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	storedToken, exists := c.tokens[token]
	if !exists {
		return false
	}

	if time.Now().After(storedToken.ExpiresAt) {
		// Token expired
		go func() {
			c.mutex.Lock()
			delete(c.tokens, token)
			c.mutex.Unlock()
		}()

		return false
	}

	return true
}

func (c *CSRFManager) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for token, csrfToken := range c.tokens {
			if now.After(csrfToken.ExpiresAt) {
				delete(c.tokens, token)
			}
		}
		c.mutex.Unlock()
	}
}

// CSRF middleware for protecting state-changing operations
func (s *Server) csrfProtection(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Only protect state-changing operations
		if request.Method == httpMethodPOST || request.Method == httpMethodPUT || request.Method == httpMethodDELETE {
			token := request.Header.Get("X-CSRF-Token")
			if token == "" {
				token = request.FormValue("csrf_token")
			}

			if token == "" || !s.csrfManager.ValidateToken(token) {
				http.Error(writer, "CSRF token missing or invalid", http.StatusForbidden)

				return
			}
		}

		next(writer, request)
	}
}

// Helper to get CSRF token for templates
func (s *Server) getCSRFToken() string {
	token, err := s.csrfManager.GenerateToken()
	if err != nil {
		// Log error but don't break the app

		return ""
	}

	return token
}
