package worker_test

import (
	"errors"
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
