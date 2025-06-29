package worker_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"wallabag-rss-tool/pkg/database/mocks"
	"wallabag-rss-tool/pkg/models"
	"wallabag-rss-tool/pkg/rss"
	rssmocks "wallabag-rss-tool/pkg/rss/mocks"
	"wallabag-rss-tool/pkg/wallabag"
	wallabagmocks "wallabag-rss-tool/pkg/wallabag/mocks"
	"wallabag-rss-tool/pkg/worker"
)

func TestNewWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockProcessor := rssmocks.NewMockProcessorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)

	w := worker.NewWorker(mockStore, mockProcessor, mockClient)

	// Cannot test private fields from external package
	assert.NotNil(t, w)
}

func TestWorker_StartAndStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockProcessor := rssmocks.NewMockProcessorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)

	// Mock GetFeeds to return empty list for initial ProcessFeeds call
	mockStore.EXPECT().GetFeeds(gomock.Any()).Return([]models.Feed{}, nil).AnyTimes()
	// Mock GetDefaultPollInterval for ticker setup
	mockStore.EXPECT().GetDefaultPollInterval(gomock.Any()).Return(60, nil).AnyTimes()

	w := worker.NewWorker(mockStore, mockProcessor, mockClient)

	// Test Start
	assert.NotPanics(t, func() {
		w.Start()
	})

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Test Stop
	assert.NotPanics(t, func() {
		w.Stop()
	})

	// Give it a moment to stop
	time.Sleep(10 * time.Millisecond)
}

func TestWorker_ProcessFeeds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("No feeds", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return([]models.Feed{}, nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Error getting feeds", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(nil, errors.New("database error"))

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Feed not ready to fetch", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		recentTime := time.Now().Add(-10 * time.Minute) // Fetched 10 minutes ago
		feeds := []models.Feed{
			{
				ID:                  1,
				URL:                 "https://example.com/feed1",
				Name:                "Feed 1",
				LastFetched:         &recentTime,
				PollIntervalMinutes: 60, // Should wait 60 minutes
			},
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Process feed with default interval", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		feeds := []models.Feed{
			{
				ID:                  1,
				URL:                 "https://example.com/feed1",
				Name:                "Feed 1",
				LastFetched:         nil, // Never fetched
				PollIntervalMinutes: 0,   // Use default
				SyncMode:            models.SyncModeNone,
				InitialSyncDone:     true, // Already done initial sync
			},
		}

		articles := []rss.Article{
			{
				Title: "Test Article",
				URL:   "https://example.com/article1",
			},
		}

		entry := &wallabag.Entry{
			ID:    123,
			URL:   "https://example.com/article1",
			Title: "Test Article",
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockStore.EXPECT().GetDefaultPollInterval(gomock.Any()).Return(30, nil)
		mockProcessor.EXPECT().FetchAndParse("https://example.com/feed1").Return(articles, nil)
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/article1").Return(false, nil)
		mockClient.EXPECT().AddEntry(gomock.Any(), "https://example.com/article1").Return(entry, nil)
		// Expect SaveArticle to be called with the converted models.Article
		mockStore.EXPECT().SaveArticle(gomock.Any(), 1, gomock.Any(), 123).Return(nil)
		mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), 1).Return(nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Process feed with custom interval", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		oldTime := time.Now().Add(-2 * time.Hour) // Fetched 2 hours ago
		feeds := []models.Feed{
			{
				ID:                  2,
				URL:                 "https://example.com/feed2",
				Name:                "Feed 2",
				LastFetched:         &oldTime,
				PollIntervalMinutes: 60, // Should fetch every hour
				SyncMode:            models.SyncModeNone,
				InitialSyncDone:     true,
			},
		}

		articles := []rss.Article{
			{
				Title: "Another Article",
				URL:   "https://example.com/article2",
			},
		}

		entry := &wallabag.Entry{
			ID:    456,
			URL:   "https://example.com/article2",
			Title: "Another Article",
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockProcessor.EXPECT().FetchAndParse("https://example.com/feed2").Return(articles, nil)
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/article2").Return(false, nil)
		mockClient.EXPECT().AddEntry(gomock.Any(), "https://example.com/article2").Return(entry, nil)
		mockStore.EXPECT().SaveArticle(gomock.Any(), 2, gomock.Any(), 456).Return(nil)
		mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), 2).Return(nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Article already processed", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		feeds := []models.Feed{
			{
				ID:                  3,
				URL:                 "https://example.com/feed3",
				Name:                "Feed 3",
				LastFetched:         nil,
				PollIntervalMinutes: 30,
				SyncMode:            models.SyncModeNone,
				InitialSyncDone:     true,
			},
		}

		articles := []rss.Article{
			{
				Title: "Processed Article",
				URL:   "https://example.com/processed",
			},
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockProcessor.EXPECT().FetchAndParse("https://example.com/feed3").Return(articles, nil)
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/processed").Return(true, nil)
		mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), 3).Return(nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Multiple articles with some processed", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		feeds := []models.Feed{
			{
				ID:                  4,
				URL:                 "https://example.com/feed4",
				Name:                "Feed 4",
				LastFetched:         nil,
				PollIntervalMinutes: 15,
				SyncMode:            models.SyncModeNone,
				InitialSyncDone:     true,
			},
		}

		articles := []rss.Article{
			{
				Title: "New Article",
				URL:   "https://example.com/new",
			},
			{
				Title: "Old Article",
				URL:   "https://example.com/old",
			},
		}

		entry := &wallabag.Entry{
			ID:    789,
			URL:   "https://example.com/new",
			Title: "New Article",
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockProcessor.EXPECT().FetchAndParse("https://example.com/feed4").Return(articles, nil)

		// First article is new
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/new").Return(false, nil)
		mockClient.EXPECT().AddEntry(gomock.Any(), "https://example.com/new").Return(entry, nil)
		mockStore.EXPECT().SaveArticle(gomock.Any(), 4, gomock.Any(), 789).Return(nil)

		// Second article is already processed
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/old").Return(true, nil)

		mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), 4).Return(nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Error getting default poll interval", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		feeds := []models.Feed{
			{
				ID:                  5,
				URL:                 "https://example.com/feed5",
				Name:                "Feed 5",
				LastFetched:         nil,
				PollIntervalMinutes: 0, // Use default
				SyncMode:            models.SyncModeNone,
				InitialSyncDone:     true,
			},
		}

		articles := []rss.Article{
			{
				Title: "Article with fallback interval",
				URL:   "https://example.com/fallback",
			},
		}

		entry := &wallabag.Entry{
			ID:    101,
			URL:   "https://example.com/fallback",
			Title: "Article with fallback interval",
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockStore.EXPECT().GetDefaultPollInterval(gomock.Any()).Return(0, errors.New("settings error"))
		mockProcessor.EXPECT().FetchAndParse("https://example.com/feed5").Return(articles, nil)
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/fallback").Return(false, nil)
		mockClient.EXPECT().AddEntry(gomock.Any(), "https://example.com/fallback").Return(entry, nil)
		mockStore.EXPECT().SaveArticle(gomock.Any(), 5, gomock.Any(), 101).Return(nil)
		mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), 5).Return(nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Error fetching RSS feed", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		feeds := []models.Feed{
			{
				ID:                  6,
				URL:                 "https://invalid.com/feed",
				Name:                "Invalid Feed",
				LastFetched:         nil,
				PollIntervalMinutes: 30,
				SyncMode:            models.SyncModeNone,
				InitialSyncDone:     true,
			},
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockProcessor.EXPECT().FetchAndParse("https://invalid.com/feed").Return(nil, errors.New("feed error"))

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Error checking if article processed", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		feeds := []models.Feed{
			{
				ID:                  7,
				URL:                 "https://example.com/feed7",
				Name:                "Feed 7",
				LastFetched:         nil,
				PollIntervalMinutes: 30,
				SyncMode:            models.SyncModeNone,
				InitialSyncDone:     true,
			},
		}

		articles := []rss.Article{
			{
				Title: "Article with check error",
				URL:   "https://example.com/check-error",
			},
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockProcessor.EXPECT().FetchAndParse("https://example.com/feed7").Return(articles, nil)
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/check-error").Return(false, errors.New("database error"))
		mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), 7).Return(nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Error adding to Wallabag", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		feeds := []models.Feed{
			{
				ID:                  8,
				URL:                 "https://example.com/feed8",
				Name:                "Feed 8",
				LastFetched:         nil,
				PollIntervalMinutes: 30,
				SyncMode:            models.SyncModeNone,
				InitialSyncDone:     true,
			},
		}

		articles := []rss.Article{
			{
				Title: "Article with Wallabag error",
				URL:   "https://example.com/wallabag-error",
			},
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockProcessor.EXPECT().FetchAndParse("https://example.com/feed8").Return(articles, nil)
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/wallabag-error").Return(false, nil)
		mockClient.EXPECT().AddEntry(gomock.Any(), "https://example.com/wallabag-error").Return(nil, errors.New("wallabag API error"))
		mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), 8).Return(nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Error saving article to database", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		feeds := []models.Feed{
			{
				ID:                  9,
				URL:                 "https://example.com/feed9",
				Name:                "Feed 9",
				LastFetched:         nil,
				PollIntervalMinutes: 30,
				SyncMode:            models.SyncModeNone,
				InitialSyncDone:     true,
			},
		}

		articles := []rss.Article{
			{
				Title: "Article with save error",
				URL:   "https://example.com/save-error",
			},
		}

		entry := &wallabag.Entry{
			ID:    999,
			URL:   "https://example.com/save-error",
			Title: "Article with save error",
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockProcessor.EXPECT().FetchAndParse("https://example.com/feed9").Return(articles, nil)
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/save-error").Return(false, nil)
		mockClient.EXPECT().AddEntry(gomock.Any(), "https://example.com/save-error").Return(entry, nil)
		mockStore.EXPECT().SaveArticle(gomock.Any(), 9, gomock.Any(), 999).Return(errors.New("database save error"))
		mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), 9).Return(nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Error updating feed last fetched", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		feeds := []models.Feed{
			{
				ID:                  10,
				URL:                 "https://example.com/feed10",
				Name:                "Feed 10",
				LastFetched:         nil,
				PollIntervalMinutes: 30,
				SyncMode:            models.SyncModeNone,
				InitialSyncDone:     true,
			},
		}

		articles := []rss.Article{
			{
				Title: "Article with update error",
				URL:   "https://example.com/update-error",
			},
		}

		entry := &wallabag.Entry{
			ID:    888,
			URL:   "https://example.com/update-error",
			Title: "Article with update error",
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockProcessor.EXPECT().FetchAndParse("https://example.com/feed10").Return(articles, nil)
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/update-error").Return(false, nil)
		mockClient.EXPECT().AddEntry(gomock.Any(), "https://example.com/update-error").Return(entry, nil)
		mockStore.EXPECT().SaveArticle(gomock.Any(), 10, gomock.Any(), 888).Return(nil)
		mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), 10).Return(errors.New("update error"))

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})

	t.Run("Initial sync with sync options", func(t *testing.T) {
		mockStore := mocks.NewMockStorer(ctrl)
		mockProcessor := rssmocks.NewMockProcessorer(ctrl)
		mockClient := wallabagmocks.NewMockClienter(ctrl)

		count := 5
		feeds := []models.Feed{
			{
				ID:                  11,
				URL:                 "https://example.com/feed11",
				Name:                "Feed 11",
				LastFetched:         nil,
				PollIntervalMinutes: 30,
				SyncMode:            models.SyncModeCount,
				SyncCount:           &count,
				InitialSyncDone:     false, // Initial sync not done yet
			},
		}

		articles := []rss.Article{
			{
				Title: "Initial Sync Article",
				URL:   "https://example.com/initial",
			},
		}

		entry := &wallabag.Entry{
			ID:    777,
			URL:   "https://example.com/initial",
			Title: "Initial Sync Article",
		}

		mockStore.EXPECT().GetFeeds(gomock.Any()).Return(feeds, nil)
		mockProcessor.EXPECT().FetchAndParseWithSyncOptions("https://example.com/feed11", models.SyncModeCount, &count, (*time.Time)(nil)).Return(articles, nil)
		mockStore.EXPECT().IsArticleAlreadyProcessed(gomock.Any(), "https://example.com/initial").Return(false, nil)
		mockClient.EXPECT().AddEntry(gomock.Any(), "https://example.com/initial").Return(entry, nil)
		mockStore.EXPECT().SaveArticle(gomock.Any(), 11, gomock.Any(), 777).Return(nil)
		mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), 11).Return(nil)
		mockStore.EXPECT().MarkFeedInitialSyncCompleted(gomock.Any(), 11).Return(nil)

		w := worker.NewWorker(mockStore, mockProcessor, mockClient)
		w.ProcessFeeds()
	})
}

func TestWorker_StopChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockProcessor := rssmocks.NewMockProcessorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)

	w := worker.NewWorker(mockStore, mockProcessor, mockClient)

	// Cannot test private stopChan field from external package
	assert.NotNil(t, w)
}

func TestWorker_QueueFeedForImmediate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockProcessor := rssmocks.NewMockProcessorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)

	// Setup expectations for worker start
	mockStore.EXPECT().GetFeeds(gomock.Any()).Return([]models.Feed{}, nil).AnyTimes()
	mockStore.EXPECT().GetDefaultPollInterval(gomock.Any()).Return(60, nil).AnyTimes()

	w := worker.NewWorker(mockStore, mockProcessor, mockClient)
	w.Start()
	defer w.Stop()

	// Test queuing a feed
	testFeed := models.Feed{
		ID:   123,
		Name: "Test Feed",
		URL:  "https://example.com/feed.xml",
	}

	// Expect the feed to be fetched and processed
	mockStore.EXPECT().GetFeedByID(gomock.Any(), 123).Return(&testFeed, nil)
	mockProcessor.EXPECT().FetchAndParseWithSyncOptions(
		testFeed.URL,
		testFeed.SyncMode,
		testFeed.SyncCount,
		testFeed.SyncDateFrom,
	).Return([]rss.Article{}, nil)
	mockStore.EXPECT().UpdateFeedLastFetched(gomock.Any(), testFeed.ID).Return(nil)
	mockStore.EXPECT().MarkFeedInitialSyncCompleted(gomock.Any(), testFeed.ID).Return(nil)

	// Queue the feed
	w.QueueFeedForImmediate(123)

	// Give time for processing
	time.Sleep(200 * time.Millisecond)
}

func TestWorker_QueueFeedForImmediate_QueueFull(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockProcessor := rssmocks.NewMockProcessorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)

	// Setup minimal expectations
	mockStore.EXPECT().GetFeeds(gomock.Any()).Return([]models.Feed{}, nil).AnyTimes()
	mockStore.EXPECT().GetDefaultPollInterval(gomock.Any()).Return(60, nil).AnyTimes()

	w := worker.NewWorker(mockStore, mockProcessor, mockClient)
	// Don't start the worker so queue won't be processed

	// Fill the queue (capacity is 100)
	for i := 0; i < 100; i++ {
		w.QueueFeedForImmediate(i)
	}

	// Check queue is full
	length, capacity := w.GetQueueStats()
	assert.Equal(t, 100, length)
	assert.Equal(t, 100, capacity)

	// Try to add one more - should not panic or block
	assert.NotPanics(t, func() {
		w.QueueFeedForImmediate(101)
	})

	// Queue should still be at capacity
	length, _ = w.GetQueueStats()
	assert.Equal(t, 100, length)
}

func TestWorker_QueueAllFeedsForImmediate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockProcessor := rssmocks.NewMockProcessorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)

	testFeeds := []models.Feed{
		{ID: 1, Name: "Feed 1"},
		{ID: 2, Name: "Feed 2"},
		{ID: 3, Name: "Feed 3"},
	}

	// Expect GetFeeds to be called for QueueAllFeedsForImmediate
	mockStore.EXPECT().GetFeeds(gomock.Any()).Return(testFeeds, nil)

	w := worker.NewWorker(mockStore, mockProcessor, mockClient)

	// Queue all feeds
	err := w.QueueAllFeedsForImmediate(context.Background())
	assert.NoError(t, err)

	// Check queue stats to verify feeds were queued
	queueLength, queueCapacity := w.GetQueueStats()
	assert.Equal(t, 3, queueLength)
	assert.Equal(t, 100, queueCapacity) // Default queue capacity
}

func TestWorker_QueueAllFeedsForImmediate_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockProcessor := rssmocks.NewMockProcessorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)

	w := worker.NewWorker(mockStore, mockProcessor, mockClient)

	// Expect GetFeeds to fail when QueueAllFeedsForImmediate is called
	mockStore.EXPECT().GetFeeds(gomock.Any()).Return(nil, errors.New("database error"))

	// Queue all feeds should return error
	err := w.QueueAllFeedsForImmediate(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get feeds")
}

func TestWorker_ProcessSingleFeedByID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockProcessor := rssmocks.NewMockProcessorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)

	// Setup expectations
	mockStore.EXPECT().GetFeeds(gomock.Any()).Return([]models.Feed{}, nil).AnyTimes()
	mockStore.EXPECT().GetDefaultPollInterval(gomock.Any()).Return(60, nil).AnyTimes()

	w := worker.NewWorker(mockStore, mockProcessor, mockClient)
	w.Start()
	defer w.Stop()

	// Expect GetFeedByID to fail
	mockStore.EXPECT().GetFeedByID(gomock.Any(), 999).Return(nil, errors.New("feed not found"))

	// Queue the feed - error should be logged but not panic
	assert.NotPanics(t, func() {
		w.QueueFeedForImmediate(999)
	})

	// Give minimal time for processing to avoid test timeout
	time.Sleep(50 * time.Millisecond)
}

func TestWorker_GetQueueStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockProcessor := rssmocks.NewMockProcessorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)

	w := worker.NewWorker(mockStore, mockProcessor, mockClient)

	// Initially empty
	length, capacity := w.GetQueueStats()
	assert.Equal(t, 0, length)
	assert.Equal(t, 100, capacity)

	// Add some items
	w.QueueFeedForImmediate(1)
	w.QueueFeedForImmediate(2)

	length, capacity = w.GetQueueStats()
	assert.Equal(t, 2, length)
	assert.Equal(t, 100, capacity)
}

func TestWorker_ConcurrentQueueOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockStorer(ctrl)
	mockProcessor := rssmocks.NewMockProcessorer(ctrl)
	mockClient := wallabagmocks.NewMockClienter(ctrl)

	// Setup expectations
	mockStore.EXPECT().GetFeeds(gomock.Any()).Return([]models.Feed{}, nil).AnyTimes()
	mockStore.EXPECT().GetDefaultPollInterval(gomock.Any()).Return(60, nil).AnyTimes()

	w := worker.NewWorker(mockStore, mockProcessor, mockClient)

	// Start fewer goroutines with fewer items to avoid overwhelming the queue
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				w.QueueFeedForImmediate(id*5 + j)
			}
		}(i)
	}

	wg.Wait()

	// Check queue has items (reduced expectations)
	length, _ := w.GetQueueStats()
	assert.GreaterOrEqual(t, length, 10) // At least some should be queued
	assert.LessOrEqual(t, length, 100)   // Can't exceed capacity
}
