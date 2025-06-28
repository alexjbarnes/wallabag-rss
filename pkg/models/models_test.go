package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"wallabag-rss-tool/pkg/models"
)

func TestFeedStruct(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		feed models.Feed
	}{
		{
			name: "Complete feed",
			feed: models.Feed{
				ID:                  1,
				URL:                 "https://example.com/feed.rss",
				Name:                "Example Feed",
				LastFetched:         &now,
				PollIntervalMinutes: 60,
			},
		},
		{
			name: "Feed without last fetched",
			feed: models.Feed{
				ID:                  2,
				URL:                 "https://test.com/rss",
				Name:                "Test Feed",
				LastFetched:         nil,
				PollIntervalMinutes: 30,
			},
		},
		{
			name: "Zero values",
			feed: models.Feed{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that struct can be created and accessed
			feed := tt.feed

			assert.Equal(t, tt.feed.ID, feed.ID)
			assert.Equal(t, tt.feed.URL, feed.URL)
			assert.Equal(t, tt.feed.Name, feed.Name)
			assert.Equal(t, tt.feed.LastFetched, feed.LastFetched)
			assert.Equal(t, tt.feed.PollIntervalMinutes, feed.PollIntervalMinutes)
		})
	}
}

func TestArticleStruct(t *testing.T) {
	now := time.Now()
	published := time.Now().Add(-time.Hour)
	wallabagID := 123

	tests := []struct {
		name    string
		article models.Article
	}{
		{
			name: "Complete article",
			article: models.Article{
				ID:              1,
				FeedID:          10,
				Title:           "Test Article",
				URL:             "https://example.com/article1",
				WallabagEntryID: &wallabagID,
				PublishedAt:     &published,
				CreatedAt:       now,
			},
		},
		{
			name: "Article without wallabag ID",
			article: models.Article{
				ID:              2,
				FeedID:          20,
				Title:           "Another Article",
				URL:             "https://test.com/article2",
				WallabagEntryID: nil,
				PublishedAt:     &published,
				CreatedAt:       now,
			},
		},
		{
			name: "Article without published date",
			article: models.Article{
				ID:              3,
				FeedID:          30,
				Title:           "Recent Article",
				URL:             "https://news.com/article3",
				WallabagEntryID: &wallabagID,
				PublishedAt:     nil,
				CreatedAt:       now,
			},
		},
		{
			name:    "Zero values",
			article: models.Article{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that struct can be created and accessed
			article := tt.article

			assert.Equal(t, tt.article.ID, article.ID)
			assert.Equal(t, tt.article.FeedID, article.FeedID)
			assert.Equal(t, tt.article.Title, article.Title)
			assert.Equal(t, tt.article.URL, article.URL)
			assert.Equal(t, tt.article.WallabagEntryID, article.WallabagEntryID)
			assert.Equal(t, tt.article.PublishedAt, article.PublishedAt)
			assert.Equal(t, tt.article.CreatedAt, article.CreatedAt)
		})
	}
}

func TestFeedPointerFields(t *testing.T) {
	t.Run("LastFetched pointer behavior", func(t *testing.T) {
		now := time.Now()

		// Test nil pointer
		feed1 := models.Feed{LastFetched: nil}
		assert.Nil(t, feed1.LastFetched)

		// Test non-nil pointer
		feed2 := models.Feed{LastFetched: &now}
		assert.NotNil(t, feed2.LastFetched)
		assert.Equal(t, now, *feed2.LastFetched)

		// Test pointer assignment
		feed3 := models.Feed{}
		feed3.LastFetched = &now
		assert.NotNil(t, feed3.LastFetched)
		assert.Equal(t, now, *feed3.LastFetched)
	})
}

func TestArticlePointerFields(t *testing.T) {
	t.Run("WallabagEntryID pointer behavior", func(t *testing.T) {
		wallabagID := 456

		// Test nil pointer
		article1 := models.Article{WallabagEntryID: nil}
		assert.Nil(t, article1.WallabagEntryID)

		// Test non-nil pointer
		article2 := models.Article{WallabagEntryID: &wallabagID}
		assert.NotNil(t, article2.WallabagEntryID)
		assert.Equal(t, wallabagID, *article2.WallabagEntryID)
	})

	t.Run("PublishedAt pointer behavior", func(t *testing.T) {
		published := time.Now().Add(-2 * time.Hour)

		// Test nil pointer
		article1 := models.Article{PublishedAt: nil}
		assert.Nil(t, article1.PublishedAt)

		// Test non-nil pointer
		article2 := models.Article{PublishedAt: &published}
		assert.NotNil(t, article2.PublishedAt)
		assert.Equal(t, published, *article2.PublishedAt)
	})
}

func TestStructCopy(t *testing.T) {
	t.Run("Feed copy", func(t *testing.T) {
		now := time.Now()
		original := models.Feed{
			ID:                  1,
			URL:                 "https://example.com/feed",
			Name:                "Original Feed",
			LastFetched:         &now,
			PollIntervalMinutes: 60,
		}

		copied := original
		assert.Equal(t, original.ID, copied.ID)
		assert.Equal(t, original.URL, copied.URL)
		assert.Equal(t, original.Name, copied.Name)
		assert.Equal(t, original.PollIntervalMinutes, copied.PollIntervalMinutes)

		// Pointer should point to same time value
		assert.Equal(t, original.LastFetched, copied.LastFetched)
		if original.LastFetched != nil && copied.LastFetched != nil {
			assert.Equal(t, *original.LastFetched, *copied.LastFetched)
		}
	})

	t.Run("Article copy", func(t *testing.T) {
		now := time.Now()
		published := time.Now().Add(-time.Hour)
		wallabagID := 789

		original := models.Article{
			ID:              1,
			FeedID:          10,
			Title:           "Original Article",
			URL:             "https://example.com/article",
			WallabagEntryID: &wallabagID,
			PublishedAt:     &published,
			CreatedAt:       now,
		}

		copied := original
		assert.Equal(t, original.ID, copied.ID)
		assert.Equal(t, original.FeedID, copied.FeedID)
		assert.Equal(t, original.Title, copied.Title)
		assert.Equal(t, original.URL, copied.URL)
		assert.Equal(t, original.CreatedAt, copied.CreatedAt)

		// Pointers should point to same values
		assert.Equal(t, original.WallabagEntryID, copied.WallabagEntryID)
		assert.Equal(t, original.PublishedAt, copied.PublishedAt)
		if original.WallabagEntryID != nil && copied.WallabagEntryID != nil {
			assert.Equal(t, *original.WallabagEntryID, *copied.WallabagEntryID)
		}
		if original.PublishedAt != nil && copied.PublishedAt != nil {
			assert.Equal(t, *original.PublishedAt, *copied.PublishedAt)
		}
	})
}
