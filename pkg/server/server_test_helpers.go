package server

import (
	"context"
	"net/http"
)

// Test helpers to access unexported server functionality

// TestGetDefaultPollIntervalWithFallback tests the getDefaultPollIntervalWithFallback method
func TestGetDefaultPollIntervalWithFallback(ctx context.Context, s *Server) int {
	return s.getDefaultPollIntervalWithFallback(ctx)
}

// TestExtractFeedIDFromPath tests the ExtractFeedIDFromPath method
func TestExtractFeedIDFromPath(s *Server, path string) (int, error) {
	return s.ExtractFeedIDFromPath(path)
}

// TestHandleIndex tests the HandleIndex method
func TestHandleIndex(s *Server) http.HandlerFunc {
	return s.HandleIndex
}

// TestHandleFeeds tests the handleFeeds method
func TestHandleFeeds(s *Server) http.HandlerFunc {
	return s.handleFeeds
}

// TestHandleFeedsGet tests the handleFeedsGet method
func TestHandleFeedsGet(s *Server) http.HandlerFunc {
	return s.handleFeedsGet
}

// TestHandleFeedsPost tests the handleFeedsPost method
func TestHandleFeedsPost(s *Server) http.HandlerFunc {
	return s.handleFeedsPost
}

// TestHandleFeedsPut tests the handleFeedsPut method
func TestHandleFeedsPut(s *Server) http.HandlerFunc {
	return s.handleFeedsPut
}

// TestHandleFeedsDelete tests the handleFeedsDelete method
func TestHandleFeedsDelete(s *Server) http.HandlerFunc {
	return s.handleFeedsDelete
}

// SetCSRFManager sets the CSRF manager for testing
func (s *Server) SetCSRFManager(manager *CSRFManager) {
	s.csrfManager = manager
}