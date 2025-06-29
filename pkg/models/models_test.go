package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"wallabag-rss-tool/pkg/models"
)

func TestSyncMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     models.SyncMode
		expected string
	}{
		{"models.SyncModeNone", models.SyncModeNone, "none"},
		{"models.SyncModeAll", models.SyncModeAll, "all"},
		{"models.SyncModeCount", models.SyncModeCount, "count"},
		{"models.SyncModeDateFrom", models.SyncModeDateFrom, "date_from"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, models.SyncMode(tt.expected), tt.mode)
			assert.Equal(t, tt.expected, string(tt.mode))
		})
	}
}

func TestTimeUnit(t *testing.T) {
	tests := []struct {
		name     string
		unit     models.TimeUnit
		expected string
	}{
		{"models.TimeUnitMinutes", models.TimeUnitMinutes, "minutes"},
		{"models.TimeUnitHours", models.TimeUnitHours, "hours"},
		{"models.TimeUnitDays", models.TimeUnitDays, "days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, models.TimeUnit(tt.expected), tt.unit)
			assert.Equal(t, tt.expected, string(tt.unit))
		})
	}
}

func TestFeed_GetPollIntervalMinutes(t *testing.T) {
	tests := []struct {
		name     string
		feed     models.Feed
		expected int
	}{
		{
			name: "poll interval in minutes",
			feed: models.Feed{
				PollInterval:     30,
				PollIntervalUnit: models.TimeUnitMinutes,
			},
			expected: 30,
		},
		{
			name: "poll interval in hours",
			feed: models.Feed{
				PollInterval:     2,
				PollIntervalUnit: models.TimeUnitHours,
			},
			expected: 120, // 2 hours = 120 minutes
		},
		{
			name: "poll interval in days",
			feed: models.Feed{
				PollInterval:     1,
				PollIntervalUnit: models.TimeUnitDays,
			},
			expected: 1440, // 1 day = 1440 minutes
		},
		{
			name: "poll interval with zero value",
			feed: models.Feed{
				PollInterval:     0,
				PollIntervalUnit: models.TimeUnitHours,
			},
			expected: 0,
		},
		{
			name: "poll interval with negative value",
			feed: models.Feed{
				PollInterval:     -5,
				PollIntervalUnit: models.TimeUnitHours,
			},
			expected: 0,
		},
		{
			name: "poll interval with unknown unit",
			feed: models.Feed{
				PollInterval:     60,
				PollIntervalUnit: models.TimeUnit("unknown"),
			},
			expected: 60, // fallback to treating as minutes
		},
		{
			name: "poll interval with empty unit",
			feed: models.Feed{
				PollInterval:     45,
				PollIntervalUnit: models.TimeUnit(""),
			},
			expected: 45, // fallback to treating as minutes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.feed.GetPollIntervalMinutes())
		})
	}
}

func TestFeed_SetPollInterval(t *testing.T) {
	tests := []struct {
		expectedUnit         models.TimeUnit
		unit                 models.TimeUnit
		name                 string
		expectedInterval     int
		expectedMinutes      int
		interval             int
	}{
		{
			name:             "set poll interval in minutes",
			interval:         30,
			unit:             models.TimeUnitMinutes,
			expectedInterval: 30,
			expectedUnit:     models.TimeUnitMinutes,
			expectedMinutes:  30,
		},
		{
			name:             "set poll interval in hours",
			interval:         3,
			unit:             models.TimeUnitHours,
			expectedInterval: 3,
			expectedUnit:     models.TimeUnitHours,
			expectedMinutes:  180, // 3 hours = 180 minutes
		},
		{
			name:             "set poll interval in days",
			interval:         2,
			unit:             models.TimeUnitDays,
			expectedInterval: 2,
			expectedUnit:     models.TimeUnitDays,
			expectedMinutes:  2880, // 2 days = 2880 minutes
		},
		{
			name:             "set poll interval with zero value",
			interval:         0,
			unit:             models.TimeUnitHours,
			expectedInterval: 0,
			expectedUnit:     models.TimeUnitHours,
			expectedMinutes:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed := &models.Feed{}
			feed.SetPollInterval(tt.interval, tt.unit)
			
			assert.Equal(t, tt.expectedInterval, feed.PollInterval)
			assert.Equal(t, tt.expectedUnit, feed.PollIntervalUnit)
			assert.Equal(t, tt.expectedMinutes, feed.PollIntervalMinutes)
		})
	}

	t.Run("set poll interval overwrites existing values", func(t *testing.T) {
		feed := &models.Feed{
			PollInterval:        60,
			PollIntervalUnit:    models.TimeUnitMinutes,
			PollIntervalMinutes: 60,
		}
		
		feed.SetPollInterval(5, models.TimeUnitHours)
		
		assert.Equal(t, 5, feed.PollInterval)
		assert.Equal(t, models.TimeUnitHours, feed.PollIntervalUnit)
		assert.Equal(t, 300, feed.PollIntervalMinutes) // 5 hours = 300 minutes
	})
}

func TestFeedStruct(t *testing.T) {
	now := time.Now()
	syncCount := 10

	tests := []struct {
		checkFunc func(t *testing.T, feed models.Feed)
		name      string
		feed      models.Feed
	}{
		{
			name: "Complete feed with all fields",
			feed: models.Feed{
				ID:                  1,
				URL:                 "https://example.com/feed.rss",
				Name:                "Example models.Feed",
				LastFetched:         &now,
				PollInterval:        2,
				PollIntervalUnit:    models.TimeUnitHours,
				PollIntervalMinutes: 120,
				SyncMode:            models.SyncModeCount,
				SyncCount:           &syncCount,
				SyncDateFrom:        &now,
				InitialSyncDone:     true,
			},
			checkFunc: func(t *testing.T, feed models.Feed) {
				t.Helper()
				assert.Equal(t, 1, feed.ID)
				assert.Equal(t, "https://example.com/feed.rss", feed.URL)
				assert.Equal(t, "Example models.Feed", feed.Name)
				assert.Equal(t, &now, feed.LastFetched)
				assert.Equal(t, 2, feed.PollInterval)
				assert.Equal(t, models.TimeUnitHours, feed.PollIntervalUnit)
				assert.Equal(t, 120, feed.PollIntervalMinutes)
				assert.Equal(t, models.SyncModeCount, feed.SyncMode)
				assert.Equal(t, &syncCount, feed.SyncCount)
				assert.Equal(t, &now, feed.SyncDateFrom)
				assert.True(t, feed.InitialSyncDone)
			},
		},
		{
			name: "models.Feed with nil pointers",
			feed: models.Feed{
				ID:               2,
				URL:              "https://test.com/rss",
				Name:             "Test models.Feed",
				LastFetched:      nil,
				SyncCount:        nil,
				SyncDateFrom:     nil,
				InitialSyncDone:  false,
			},
			checkFunc: func(t *testing.T, feed models.Feed) {
				t.Helper()
				assert.Equal(t, 2, feed.ID)
				assert.Equal(t, "https://test.com/rss", feed.URL)
				assert.Equal(t, "Test models.Feed", feed.Name)
				assert.Nil(t, feed.LastFetched)
				assert.Nil(t, feed.SyncCount)
				assert.Nil(t, feed.SyncDateFrom)
				assert.False(t, feed.InitialSyncDone)
			},
		},
		{
			name: "Zero values feed",
			feed: models.Feed{},
			checkFunc: func(t *testing.T, feed models.Feed) {
				t.Helper()
				assert.Equal(t, 0, feed.ID)
				assert.Equal(t, "", feed.URL)
				assert.Equal(t, "", feed.Name)
				assert.Nil(t, feed.LastFetched)
				assert.Equal(t, 0, feed.PollInterval)
				assert.Equal(t, models.TimeUnit(""), feed.PollIntervalUnit)
				assert.Equal(t, 0, feed.PollIntervalMinutes)
				assert.Equal(t, models.SyncMode(""), feed.SyncMode)
				assert.Nil(t, feed.SyncCount)
				assert.Nil(t, feed.SyncDateFrom)
				assert.False(t, feed.InitialSyncDone)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFunc(t, tt.feed)
		})
	}
}

func TestArticleStruct(t *testing.T) {
	now := time.Now()
	publishedAt := now.Add(-time.Hour)
	createdAt := now.Add(-30 * time.Minute)
	wallabagEntryID := 123

	tests := []struct {
		checkFunc func(t *testing.T, article models.Article)
		name      string
		article   models.Article
	}{
		{
			name: "Complete article with all fields",
			article: models.Article{
				ID:              1,
				FeedID:          2,
				Title:           "Test Article",
				URL:             "https://example.com/article",
				PublishedAt:     &publishedAt,
				WallabagEntryID: &wallabagEntryID,
				CreatedAt:       createdAt,
			},
			checkFunc: func(t *testing.T, article models.Article) {
				t.Helper()
				assert.Equal(t, 1, article.ID)
				assert.Equal(t, 2, article.FeedID)
				assert.Equal(t, "Test Article", article.Title)
				assert.Equal(t, "https://example.com/article", article.URL)
				assert.Equal(t, &publishedAt, article.PublishedAt)
				assert.Equal(t, &wallabagEntryID, article.WallabagEntryID)
				assert.Equal(t, createdAt, article.CreatedAt)
			},
		},
		{
			name: "models.Article with nil pointers",
			article: models.Article{
				ID:              3,
				FeedID:          4,
				Title:           "Another models.Article",
				URL:             "https://test.com/article",
				PublishedAt:     nil,
				WallabagEntryID: nil,
				CreatedAt:       now,
			},
			checkFunc: func(t *testing.T, article models.Article) {
				t.Helper()
				assert.Equal(t, 3, article.ID)
				assert.Equal(t, 4, article.FeedID)
				assert.Equal(t, "Another models.Article", article.Title)
				assert.Equal(t, "https://test.com/article", article.URL)
				assert.Nil(t, article.PublishedAt)
				assert.Nil(t, article.WallabagEntryID)
				assert.Equal(t, now, article.CreatedAt)
			},
		},
		{
			name: "Zero values article",
			article: models.Article{},
			checkFunc: func(t *testing.T, article models.Article) {
				t.Helper()
				assert.Equal(t, 0, article.ID)
				assert.Equal(t, 0, article.FeedID)
				assert.Equal(t, "", article.Title)
				assert.Equal(t, "", article.URL)
				assert.Nil(t, article.PublishedAt)
				assert.Nil(t, article.WallabagEntryID)
				assert.True(t, article.CreatedAt.IsZero())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFunc(t, tt.article)
		})
	}
}