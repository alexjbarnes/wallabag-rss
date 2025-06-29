package server

import (
	"net/http"
	"time"
)

// Test helpers to access unexported functionality

// CreateCSRFProtectedHandler creates a CSRF protected handler for testing
func CreateCSRFProtectedHandler(manager *CSRFManager, handler http.HandlerFunc) http.HandlerFunc {
	s := &Server{
		csrfManager: manager,
	}

	return s.csrfProtection(handler)
}

// CreateExpiredToken creates an expired token for testing
func CreateExpiredToken(manager *CSRFManager) string {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	
	token := "expired-test-token"
	manager.tokens[token] = CSRFToken{
		Value:     token,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	
	return token
}

// GetTokenCount returns the number of tokens stored (for testing)
func (c *CSRFManager) GetTokenCount() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.tokens)
}

// TestGetCSRFTokenHelper tests the getCSRFToken helper
func TestGetCSRFTokenHelper(manager *CSRFManager) string {
	s := &Server{
		csrfManager: manager,
	}

	return s.getCSRFToken()
}