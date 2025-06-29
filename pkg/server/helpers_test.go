package server

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"wallabag-rss-tool/pkg/database/mocks"
	"wallabag-rss-tool/pkg/rss"
	"wallabag-rss-tool/pkg/wallabag"
	wallabagmocks "wallabag-rss-tool/pkg/wallabag/mocks"
	"wallabag-rss-tool/pkg/worker"
)

func TestServerHelperFunctions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)
	
	// Create a real worker for testing (it won't be started)
	rssProcessor := rss.NewProcessor()
	realWorker := worker.NewWorker(mockStore, rssProcessor, mockClient)

	t.Run("getDefaultPollIntervalWithFallback with valid interval", func(t *testing.T) {
		// Mock successful database call
		mockStore.EXPECT().GetDefaultPollInterval(gomock.Any()).Return(120, nil)

		s := NewServer(mockStore, mockClient, realWorker)
		
		// Use helper function to test the method
		interval := TestGetDefaultPollIntervalWithFallback(context.Background(), s)
		assert.Equal(t, 120, interval)
	})

	t.Run("getDefaultPollIntervalWithFallback with database error", func(t *testing.T) {
		// Mock database error
		mockStore.EXPECT().GetDefaultPollInterval(gomock.Any()).Return(0, errors.New("database error"))

		s := NewServer(mockStore, mockClient, realWorker)
		
		// Should return fallback value (60)
		interval := TestGetDefaultPollIntervalWithFallback(context.Background(), s)
		assert.Equal(t, 60, interval)
	})

	t.Run("extractFeedIDFromPath with valid path", func(t *testing.T) {
		s := NewServer(mockStore, mockClient, realWorker)
		
		// Test valid feed ID extraction
		feedID, err := TestExtractFeedIDFromPath(s, "/feeds/123")
		assert.NoError(t, err)
		assert.Equal(t, 123, feedID)
	})

	t.Run("extractFeedIDFromPath with invalid path", func(t *testing.T) {
		s := NewServer(mockStore, mockClient, realWorker)
		
		// Test invalid paths
		_, err := TestExtractFeedIDFromPath(s, "/feeds/invalid")
		assert.Error(t, err)
		
		_, err = s.ExtractFeedIDFromPath("/feeds/")
		assert.Error(t, err)
	})
}

func TestServerHTTPHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	
	// Create real Wallabag client and worker for testing
	wallabagClient := wallabag.NewClient("http://localhost", "client", "secret", "user", "pass")
	rssProcessor := rss.NewProcessor()
	realWorker := worker.NewWorker(mockStore, rssProcessor, wallabagClient)

	t.Run("extractFeedIDFromPath helper function works correctly", func(t *testing.T) {
		s := NewServer(mockStore, wallabagClient, realWorker)
		
		// Test valid ID extraction
		id, err := s.ExtractFeedIDFromPath("/feeds/42")
		assert.NoError(t, err)
		assert.Equal(t, 42, id)
		
		// Test invalid ID
		_, err = s.ExtractFeedIDFromPath("/feeds/abc")
		assert.Error(t, err)
		
		// Test empty ID
		_, err = s.ExtractFeedIDFromPath("/feeds/")
		assert.Error(t, err)
	})
}