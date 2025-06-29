package database_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
	"wallabag-rss-tool/pkg/database"
	"wallabag-rss-tool/pkg/models"
)

//nolint:nakedret // Named returns improve readability for test helper
func setupTestDB(t *testing.T) (db *sql.DB, cleanup func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "wallabag_store_test_")
	assert.NoError(t, err)

	testDBPath := filepath.Join(tempDir, "test.db")
	db, err = sql.Open("sqlite", testDBPath)
	assert.NoError(t, err)

	// Create tables
	schema := `
CREATE TABLE feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    last_fetched DATETIME,
    poll_interval_minutes INTEGER DEFAULT 60,
    poll_interval INTEGER DEFAULT 1,
    poll_interval_unit TEXT DEFAULT 'days',
    sync_mode TEXT DEFAULT 'none',
    sync_count INTEGER,
    sync_date_from DATETIME,
    initial_sync_done BOOLEAN DEFAULT 0
);

CREATE TABLE articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    wallabag_entry_id INTEGER,
    published_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT
);

INSERT INTO settings (key, value) VALUES ('default_poll_interval_minutes', '60');
`
	_, err = db.Exec(schema)
	assert.NoError(t, err)

	cleanup = func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return
}

func TestNewSQLStore(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := database.NewSQLStore(db)
	assert.NotNil(t, store)
	// Cannot access unexported field db from external test
}

func TestSQLStore_GetFeeds(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Empty feeds table", func(t *testing.T) {
		feeds, err := store.GetFeeds(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, feeds)
	})

	t.Run("Get feeds with data", func(t *testing.T) {
		now := time.Now()

		// Insert test feeds
		_, err := db.Exec("INSERT INTO feeds (url, name, last_fetched, poll_interval_minutes, poll_interval, poll_interval_unit, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"https://example.com/feed1", "Feed 1", now, 30, 30, "minutes", "none", true)
		assert.NoError(t, err)

		_, err = db.Exec("INSERT INTO feeds (url, name, poll_interval_minutes, poll_interval, poll_interval_unit, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?, ?, ?, ?)",
			"https://example.com/feed2", "Feed 2", 60, 1, "hours", "none", false)
		assert.NoError(t, err)

		feeds, err := store.GetFeeds(context.Background())
		assert.NoError(t, err)
		assert.Len(t, feeds, 2)

		// Check first feed (with last_fetched)
		feed1 := feeds[0]
		assert.Equal(t, "https://example.com/feed1", feed1.URL)
		assert.Equal(t, "Feed 1", feed1.Name)
		assert.NotNil(t, feed1.LastFetched)
		assert.Equal(t, 30, feed1.PollIntervalMinutes)

		// Check second feed (without last_fetched)
		feed2 := feeds[1]
		assert.Equal(t, "https://example.com/feed2", feed2.URL)
		assert.Equal(t, "Feed 2", feed2.Name)
		assert.Nil(t, feed2.LastFetched)
		assert.Equal(t, 60, feed2.PollIntervalMinutes)
	})
}

func TestSQLStore_GetFeedByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Feed not found", func(t *testing.T) {
		feed, err := store.GetFeedByID(context.Background(), 999)
		assert.Error(t, err)
		assert.Nil(t, feed)
		assert.Contains(t, err.Error(), "feed with ID 999 not found")
	})

	t.Run("Get existing feed", func(t *testing.T) {
		now := time.Now()

		// Insert test feed
		res, err := db.Exec("INSERT INTO feeds (url, name, last_fetched, poll_interval_minutes, poll_interval, poll_interval_unit, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"https://example.com/feed", "Test Feed", now, 45, 45, "minutes", "none", true)
		assert.NoError(t, err)

		id, err := res.LastInsertId()
		assert.NoError(t, err)

		feed, err := store.GetFeedByID(context.Background(), int(id))
		assert.NoError(t, err)
		assert.NotNil(t, feed)
		assert.Equal(t, int(id), feed.ID)
		assert.Equal(t, "https://example.com/feed", feed.URL)
		assert.Equal(t, "Test Feed", feed.Name)
		assert.NotNil(t, feed.LastFetched)
		assert.Equal(t, 45, feed.PollIntervalMinutes)
	})

	t.Run("Get feed without last_fetched", func(t *testing.T) {
		// Insert feed without last_fetched
		res, err := db.Exec("INSERT INTO feeds (url, name, poll_interval_minutes, poll_interval, poll_interval_unit, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?, ?, ?, ?)",
			"https://example.com/feed2", "Test Feed 2", 90, 90, "minutes", "all", false)
		assert.NoError(t, err)

		id, err := res.LastInsertId()
		assert.NoError(t, err)

		feed, err := store.GetFeedByID(context.Background(), int(id))
		assert.NoError(t, err)
		assert.NotNil(t, feed)
		assert.Equal(t, int(id), feed.ID)
		assert.Equal(t, "https://example.com/feed2", feed.URL)
		assert.Equal(t, "Test Feed 2", feed.Name)
		assert.Nil(t, feed.LastFetched)
		assert.Equal(t, 90, feed.PollIntervalMinutes)
	})
}

func TestSQLStore_InsertFeed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Insert valid feed", func(t *testing.T) {
		feed := models.Feed{
			URL:                 "https://example.com/rss",
			Name:                "Example RSS",
			PollIntervalMinutes: 120,
		}

		id, err := store.InsertFeed(context.Background(), &feed)
		assert.NoError(t, err)
		assert.Greater(t, id, int64(0))

		// Verify feed was inserted
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM feeds WHERE id = ?", id).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Insert duplicate URL", func(t *testing.T) {
		feed1 := models.Feed{
			URL:                 "https://duplicate.com/rss",
			Name:                "First Feed",
			PollIntervalMinutes: 60,
		}

		feed2 := models.Feed{
			URL:                 "https://duplicate.com/rss", // Same URL
			Name:                "Second Feed",
			PollIntervalMinutes: 30,
		}

		// First insert should succeed
		_, err := store.InsertFeed(context.Background(), &feed1)
		assert.NoError(t, err)

		// Second insert should fail due to unique constraint
		_, err = store.InsertFeed(context.Background(), &feed2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to insert feed")
	})
}

func TestSQLStore_UpdateFeed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Update existing feed", func(t *testing.T) {
		// Insert initial feed
		res, err := db.Exec("INSERT INTO feeds (url, name, poll_interval_minutes, poll_interval, poll_interval_unit, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?, ?, ?, ?)",
			"https://example.com/old", "Old Name", 60, 1, "hours", "none", true)
		assert.NoError(t, err)

		id, err := res.LastInsertId()
		assert.NoError(t, err)

		// Update feed
		updatedFeed := models.Feed{
			ID:   int(id),
			URL:  "https://example.com/new",
			Name: "New Name",
		}
		updatedFeed.SetPollInterval(30, models.TimeUnitMinutes)

		err = store.UpdateFeed(context.Background(), &updatedFeed)
		assert.NoError(t, err)

		// Verify update
		var url, name string
		var interval int
		err = db.QueryRow("SELECT url, name, poll_interval_minutes FROM feeds WHERE id = ?", id).
			Scan(&url, &name, &interval)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/new", url)
		assert.Equal(t, "New Name", name)
		assert.Equal(t, 30, interval)
	})

	t.Run("Update non-existing feed", func(t *testing.T) {
		feed := models.Feed{
			ID:                  999,
			URL:                 "https://nonexistent.com",
			Name:                "Non-existent",
			PollIntervalMinutes: 60,
		}

		err := store.UpdateFeed(context.Background(), &feed)
		assert.NoError(t, err) // SQL UPDATE doesn't error when no rows are affected
	})
}

func TestSQLStore_DeleteFeed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Delete existing feed", func(t *testing.T) {
		// Insert test feed
		res, err := db.Exec("INSERT INTO feeds (url, name, poll_interval_minutes, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?, ?)",
			"https://example.com/delete", "To Delete", 60, "none", true)
		assert.NoError(t, err)

		id, err := res.LastInsertId()
		assert.NoError(t, err)

		err = store.DeleteFeed(context.Background(), int(id))
		assert.NoError(t, err)

		// Verify deletion
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM feeds WHERE id = ?", id).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Delete non-existing feed", func(t *testing.T) {
		err := store.DeleteFeed(context.Background(), 999)
		assert.NoError(t, err) // SQL DELETE doesn't error when no rows are affected
	})
}

func TestSQLStore_GetArticles(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Empty articles table", func(t *testing.T) {
		articles, err := store.GetArticles(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, articles)
	})

	t.Run("Get articles with data", func(t *testing.T) {
		// Insert test feed first
		res, err := db.Exec("INSERT INTO feeds (url, name, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?)",
			"https://example.com/feed", "Test Feed", "none", true)
		assert.NoError(t, err)
		feedID, _ := res.LastInsertId()

		now := time.Now()
		published := time.Now().Add(-time.Hour)
		wallabagID := 123

		// Insert articles
		_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, wallabag_entry_id, published_at, created_at) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			feedID, "Article 1", "https://example.com/article1", wallabagID, published, now)
		assert.NoError(t, err)

		_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, created_at) 
			VALUES (?, ?, ?, ?)`,
			feedID, "Article 2", "https://example.com/article2", now.Add(-time.Minute))
		assert.NoError(t, err)

		articles, err := store.GetArticles(context.Background())
		assert.NoError(t, err)
		assert.Len(t, articles, 2)

		// Articles should be sorted by created_at DESC
		article1 := articles[0]
		assert.Equal(t, "Article 1", article1.Title)
		assert.Equal(t, "https://example.com/article1", article1.URL)
		assert.NotNil(t, article1.WallabagEntryID)
		assert.Equal(t, wallabagID, *article1.WallabagEntryID)
		assert.NotNil(t, article1.PublishedAt)

		article2 := articles[1]
		assert.Equal(t, "Article 2", article2.Title)
		assert.Equal(t, "https://example.com/article2", article2.URL)
		assert.Nil(t, article2.WallabagEntryID)
		assert.Nil(t, article2.PublishedAt)
	})
}

func TestSQLStore_SaveArticle(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Save valid article", func(t *testing.T) {
		// Insert test feed first
		res, err := db.Exec("INSERT INTO feeds (url, name, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?)",
			"https://example.com/feed", "Test Feed", "none", true)
		assert.NoError(t, err)
		feedID, _ := res.LastInsertId()

		published := time.Now().Add(-time.Hour)
		article := models.Article{
			Title:       "Test Article",
			URL:         "https://example.com/article",
			PublishedAt: &published,
		}
		wallabagEntryID := 456

		err = store.SaveArticle(context.Background(), int(feedID), &article, wallabagEntryID)
		assert.NoError(t, err)

		// Verify article was saved
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM articles WHERE url = ?", article.URL).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify all fields
		var title, url string
		var wID int
		var pub time.Time
		err = db.QueryRow("SELECT title, url, wallabag_entry_id, published_at FROM articles WHERE url = ?",
			article.URL).Scan(&title, &url, &wID, &pub)
		assert.NoError(t, err)
		assert.Equal(t, "Test Article", title)
		assert.Equal(t, "https://example.com/article", url)
		assert.Equal(t, wallabagEntryID, wID)
	})

	t.Run("Save article with duplicate URL", func(t *testing.T) {
		// Insert test feed first
		res, err := db.Exec("INSERT INTO feeds (url, name, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?)",
			"https://example.com/feed2", "Test Feed 2", "none", true)
		assert.NoError(t, err)
		feedID, _ := res.LastInsertId()

		article := models.Article{
			Title: "Duplicate Article",
			URL:   "https://example.com/duplicate",
		}

		// First save should succeed
		err = store.SaveArticle(context.Background(), int(feedID), &article, 111)
		assert.NoError(t, err)

		// Second save should fail due to unique constraint
		err = store.SaveArticle(context.Background(), int(feedID), &article, 222)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to insert article")
	})
}

func TestSQLStore_IsArticleAlreadyProcessed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Article not processed", func(t *testing.T) {
		processed, err := store.IsArticleAlreadyProcessed(context.Background(), "https://example.com/new")
		assert.NoError(t, err)
		assert.False(t, processed)
	})

	t.Run("Article already processed", func(t *testing.T) {
		// Insert test feed and article
		res, err := db.Exec("INSERT INTO feeds (url, name, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?)",
			"https://example.com/feed", "Test Feed", "none", true)
		assert.NoError(t, err)
		feedID, _ := res.LastInsertId()

		_, err = db.Exec("INSERT INTO articles (feed_id, title, url) VALUES (?, ?, ?)",
			feedID, "Existing Article", "https://example.com/existing")
		assert.NoError(t, err)

		processed, err := store.IsArticleAlreadyProcessed(context.Background(), "https://example.com/existing")
		assert.NoError(t, err)
		assert.True(t, processed)
	})
}

func TestSQLStore_GetDefaultPollInterval(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Get existing default poll interval", func(t *testing.T) {
		interval, err := store.GetDefaultPollInterval(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 60, interval)
	})

	t.Run("Setting not found", func(t *testing.T) {
		// Remove the setting
		_, err := db.Exec("DELETE FROM settings WHERE key = ?", "default_poll_interval_minutes")
		assert.NoError(t, err)

		interval, err := store.GetDefaultPollInterval(context.Background())
		assert.Error(t, err)
		assert.Equal(t, 0, interval)
		assert.Contains(t, err.Error(), "failed to get default poll interval from settings")
	})
}

func TestSQLStore_UpdateDefaultPollInterval(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Update existing setting", func(t *testing.T) {
		err := store.UpdateDefaultPollInterval(context.Background(), 120)
		assert.NoError(t, err)

		// Verify update
		var value int
		err = db.QueryRow("SELECT value FROM settings WHERE key = ?", "default_poll_interval_minutes").Scan(&value)
		assert.NoError(t, err)
		assert.Equal(t, 120, value)
	})

	t.Run("Insert new setting", func(t *testing.T) {
		// Remove existing setting
		_, err := db.Exec("DELETE FROM settings WHERE key = ?", "default_poll_interval_minutes")
		assert.NoError(t, err)

		err = store.UpdateDefaultPollInterval(context.Background(), 180)
		assert.NoError(t, err)

		// Verify insertion
		var value int
		err = db.QueryRow("SELECT value FROM settings WHERE key = ?", "default_poll_interval_minutes").Scan(&value)
		assert.NoError(t, err)
		assert.Equal(t, 180, value)
	})
}

func TestSQLStore_UpdateFeedLastFetched(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Update existing feed", func(t *testing.T) {
		// Insert test feed
		res, err := db.Exec("INSERT INTO feeds (url, name, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?)",
			"https://example.com/feed", "Test Feed", "none", true)
		assert.NoError(t, err)
		feedID, _ := res.LastInsertId()

		beforeUpdate := time.Now()
		err = store.UpdateFeedLastFetched(context.Background(), int(feedID))
		assert.NoError(t, err)
		afterUpdate := time.Now()

		// Verify update
		var lastFetched time.Time
		err = db.QueryRow("SELECT last_fetched FROM feeds WHERE id = ?", feedID).Scan(&lastFetched)
		assert.NoError(t, err)
		assert.True(t, lastFetched.After(beforeUpdate) || lastFetched.Equal(beforeUpdate))
		assert.True(t, lastFetched.Before(afterUpdate) || lastFetched.Equal(afterUpdate))
	})

	t.Run("Update non-existing feed", func(t *testing.T) {
		err := store.UpdateFeedLastFetched(context.Background(), 999)
		assert.NoError(t, err) // SQL UPDATE doesn't error when no rows are affected
	})
}

func TestSQLStore_MarkFeedInitialSyncCompleted(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	store := database.NewSQLStore(db)

	t.Run("Mark existing feed sync completed", func(t *testing.T) {
		// Insert test feed
		res, err := db.Exec("INSERT INTO feeds (url, name, sync_mode, initial_sync_done) VALUES (?, ?, ?, ?)",
			"https://example.com/feed", "Test Feed", "none", false)
		assert.NoError(t, err)
		feedID, _ := res.LastInsertId()

		err = store.MarkFeedInitialSyncCompleted(context.Background(), int(feedID))
		assert.NoError(t, err)

		// Verify update
		var syncDone bool
		err = db.QueryRow("SELECT initial_sync_done FROM feeds WHERE id = ?", feedID).Scan(&syncDone)
		assert.NoError(t, err)
		assert.True(t, syncDone)
	})

	t.Run("Mark non-existing feed", func(t *testing.T) {
		err := store.MarkFeedInitialSyncCompleted(context.Background(), 999)
		assert.NoError(t, err) // SQL UPDATE doesn't error when no rows are affected
	})
}

func TestStore_ComprehensiveCoverage(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	store := database.NewSQLStore(db)
	ctx := context.Background()

	t.Run("Additional DeleteFeed coverage", func(t *testing.T) {
		// Create feed to delete
		feedID, err := store.InsertFeed(ctx, &models.Feed{
			Name: "Feed to Delete",
			URL:  "https://example.com/delete.xml",
		})
		assert.NoError(t, err)

		// Delete existing feed
		err = store.DeleteFeed(ctx, int(feedID))
		assert.NoError(t, err)
		
		// Verify deletion
		feed, err := store.GetFeedByID(ctx, int(feedID))
		assert.Error(t, err)
		assert.Nil(t, feed)
		
		// Delete non-existent feed (should not error)
		err = store.DeleteFeed(ctx, 99999)
		assert.NoError(t, err)
	})

	t.Run("Additional SaveArticle coverage", func(t *testing.T) {
		// Create feed for articles
		feedID, err := store.InsertFeed(ctx, &models.Feed{
			Name: "Article Test Feed",
			URL:  "https://example.com/articles.xml",
		})
		assert.NoError(t, err)

		// Save article with published date
		publishedTime := time.Now().Add(-24 * time.Hour)
		article := models.Article{
			Title:       "Article with Date",
			URL:         "https://example.com/with-date",
			PublishedAt: &publishedTime,
		}
		
		err = store.SaveArticle(ctx, int(feedID), &article, 12345)
		assert.NoError(t, err)
		
		// Save article without published date
		articleNoDate := models.Article{
			Title: "Article without Date",
			URL:   "https://example.com/no-date",
		}
		
		err = store.SaveArticle(ctx, int(feedID), &articleNoDate, 0)
		assert.NoError(t, err)
		
		// Verify articles were saved
		articles, err := store.GetArticles(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(articles), 2)
	})

	t.Run("Additional UpdateDefaultPollInterval coverage", func(t *testing.T) {
		// Test edge case values
		testValues := []int{1, 15, 60, 120, 1440, 10080} // Various intervals
		
		for _, val := range testValues {
			err := store.UpdateDefaultPollInterval(ctx, val)
			assert.NoError(t, err)
			
			// Verify the value was set
			retrieved, err := store.GetDefaultPollInterval(ctx)
			assert.NoError(t, err)
			assert.Equal(t, val, retrieved)
		}
	})

	t.Run("Additional UpdateFeedLastFetched coverage", func(t *testing.T) {
		// Create test feed
		feedID, err := store.InsertFeed(ctx, &models.Feed{
			Name: "Last Fetched Test",
			URL:  "https://example.com/lastfetch.xml",
		})
		assert.NoError(t, err)

		// Test updating last fetched timestamp
		err = store.UpdateFeedLastFetched(ctx, int(feedID))
		assert.NoError(t, err)
		
		// Test multiple updates
		err = store.UpdateFeedLastFetched(ctx, int(feedID))
		assert.NoError(t, err)
		
		err = store.UpdateFeedLastFetched(ctx, int(feedID))
		assert.NoError(t, err)
		
		// Test with invalid feed ID
		err = store.UpdateFeedLastFetched(ctx, 999999)
		assert.NoError(t, err) // Should not error
	})

	t.Run("Additional MarkFeedInitialSyncCompleted coverage", func(t *testing.T) {
		// Create test feed
		feedID, err := store.InsertFeed(ctx, &models.Feed{
			Name: "Sync Test Feed",
			URL:  "https://example.com/sync.xml",
		})
		assert.NoError(t, err)

		// Mark sync completed
		err = store.MarkFeedInitialSyncCompleted(ctx, int(feedID))
		assert.NoError(t, err)
		
		// Mark again (should not error)
		err = store.MarkFeedInitialSyncCompleted(ctx, int(feedID))
		assert.NoError(t, err)
		
		// Test with invalid feed ID
		err = store.MarkFeedInitialSyncCompleted(ctx, 999999)
		assert.NoError(t, err) // Should not error
	})

	t.Run("Additional IsArticleAlreadyProcessed coverage", func(t *testing.T) {
		// Test with various URL formats
		testURLs := []string{
			"https://example.com/article1",
			"http://example.com/article2",
			"https://sub.example.com/path/to/article",
			"https://example.com/article?param=value&other=test",
		}
		
		// Create feed for test articles
		feedID, err := store.InsertFeed(ctx, &models.Feed{
			Name: "URL Test Feed",
			URL:  "https://example.com/urltest.xml",
		})
		assert.NoError(t, err)
		
		// Save first article
		article := models.Article{
			Title: "Test Article",
			URL:   testURLs[0],
		}
		err = store.SaveArticle(ctx, int(feedID), &article, 0)
		assert.NoError(t, err)
		
		// Check first URL exists
		exists, err := store.IsArticleAlreadyProcessed(ctx, testURLs[0])
		assert.NoError(t, err)
		assert.True(t, exists)
		
		// Check other URLs don't exist
		for _, url := range testURLs[1:] {
			otherExists, err := store.IsArticleAlreadyProcessed(ctx, url)
			assert.NoError(t, err)
			assert.False(t, otherExists)
		}
		
		// Test with empty and special characters
		exists, err = store.IsArticleAlreadyProcessed(ctx, "")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestStore_EdgeCaseCoverage(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	store := database.NewSQLStore(db)
	ctx := context.Background()
	
	t.Run("SaveArticle with zero wallabag entry ID", func(t *testing.T) {
		// Create feed
		feedID, err := store.InsertFeed(ctx, &models.Feed{
			Name: "Test Feed",
			URL:  "https://example.com/test.xml",
		})
		assert.NoError(t, err)
		
		// Save article with zero wallabag entry ID
		article := models.Article{
			Title: "Article with Zero Entry ID",
			URL:   "https://example.com/zero-entry",
		}
		
		err = store.SaveArticle(ctx, int(feedID), &article, 0)
		assert.NoError(t, err)
		
		// Verify article was saved - when wallabagEntryID is 0, it's stored as 0, not NULL
		var wEntryID int
		err = db.QueryRow("SELECT wallabag_entry_id FROM articles WHERE url = ?", 
			article.URL).Scan(&wEntryID)
		assert.NoError(t, err)
		assert.Equal(t, 0, wEntryID) // Should be 0
	})
	
	t.Run("Helper functions coverage through complex operations", func(t *testing.T) {
		// Test scanFeedRow and setFeedNullableFields indirectly through GetFeeds
		now := time.Now()
		syncCount := 5
		
		// Insert a feed with all possible field combinations to test scanFeedRow
		_, err := db.Exec(`INSERT INTO feeds 
			(url, name, last_fetched, poll_interval_minutes, poll_interval, poll_interval_unit, 
			 sync_mode, sync_count, sync_date_from, initial_sync_done) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"https://example.com/complex", "Complex Feed", now, 90, 1, "hours",
			"count", syncCount, now.Add(-24*time.Hour), true)
		assert.NoError(t, err)
		
		// Fetch all feeds to trigger scanFeedRow
		feeds, err := store.GetFeeds(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(feeds), 1)
		
		// Find and verify the complex feed
		var complexFeed *models.Feed
		for _, feed := range feeds {
			if feed.Name == "Complex Feed" {
				complexFeed = &feed
				break
			}
		}
		assert.NotNil(t, complexFeed)
		assert.NotNil(t, complexFeed.LastFetched)
		assert.NotNil(t, complexFeed.SyncCount)
		assert.Equal(t, syncCount, *complexFeed.SyncCount)
		assert.Equal(t, models.SyncModeCount, complexFeed.SyncMode)
	})
	
	t.Run("Statement preparation and execution error paths", func(t *testing.T) {
		// Test error paths in functions with 70% coverage
		// Most of these functions have error handling that's hard to trigger in normal SQLite
		// but we can test the defer statement cleanup paths
		
		// Create a feed to test UpdateFeed statement preparation
		feedID, err := store.InsertFeed(ctx, &models.Feed{
			Name: "Test Feed for Statements",
			URL:  "https://example.com/statements.xml",
		})
		assert.NoError(t, err)
		
		// Test UpdateFeed with valid data (covers statement preparation and execution)
		updatedFeed := models.Feed{
			ID:   int(feedID),
			Name: "Updated Statement Test Feed",
			URL:  "https://example.com/updated-statements.xml",
		}
		updatedFeed.SetPollInterval(2, models.TimeUnitHours)
		
		err = store.UpdateFeed(ctx, &updatedFeed)
		assert.NoError(t, err)
		
		// Test DeleteFeed statement paths
		err = store.DeleteFeed(ctx, int(feedID))
		assert.NoError(t, err)
		
		// Test UpdateDefaultPollInterval statement paths with different values
		intervals := []int{30, 45, 90, 120, 180}
		for _, interval := range intervals {
			err = store.UpdateDefaultPollInterval(ctx, interval)
			assert.NoError(t, err)
		}
		
		// Test UpdateFeedLastFetched statement paths
		// Create another feed
		feedID2, err := store.InsertFeed(ctx, &models.Feed{
			Name: "Feed for Last Fetched",
			URL:  "https://example.com/lastfetched.xml",
		})
		assert.NoError(t, err)
		
		// Test multiple updates to cover statement reuse
		for i := 0; i < 3; i++ {
			err = store.UpdateFeedLastFetched(ctx, int(feedID2))
			assert.NoError(t, err)
		}
		
		// Test MarkFeedInitialSyncCompleted statement paths
		err = store.MarkFeedInitialSyncCompleted(ctx, int(feedID2))
		assert.NoError(t, err)
		
		// Test again to ensure idempotency
		err = store.MarkFeedInitialSyncCompleted(ctx, int(feedID2))
		assert.NoError(t, err)
	})
	
	t.Run("InsertFeed with complex nullable field combinations", func(t *testing.T) {
		// Test InsertFeed with different combinations of nullable fields
		testFeeds := []struct {
			syncCount   *int
			syncDate    *time.Time
			name        string
			syncMode    models.SyncMode
			initialSync bool
		}{
			{nil, nil, "Feed with nil sync count", models.SyncModeCount, false},
			{nil, nil, "Feed with nil sync date", models.SyncModeDateFrom, true},
			{nil, nil, "Feed with both nil", models.SyncModeAll, false},
		}
		
		for i, test := range testFeeds {
			feed := &models.Feed{
				Name:            test.name,
				URL:             fmt.Sprintf("https://example.com/nullable-%d.xml", i),
				SyncMode:        test.syncMode,
				SyncCount:       test.syncCount,
				SyncDateFrom:    test.syncDate,
				InitialSyncDone: test.initialSync,
			}
			feed.SetPollInterval(i+1, models.TimeUnitMinutes)
			
			feedID, err := store.InsertFeed(ctx, feed)
			assert.NoError(t, err)
			assert.Greater(t, feedID, int64(0))
			
			// Verify insertion
			retrieved, err := store.GetFeedByID(ctx, int(feedID))
			assert.NoError(t, err)
			assert.Equal(t, test.name, retrieved.Name)
			assert.Equal(t, test.syncMode, retrieved.SyncMode)
			assert.Equal(t, test.initialSync, retrieved.InitialSyncDone)
			
			if test.syncCount != nil {
				assert.NotNil(t, retrieved.SyncCount)
				assert.Equal(t, *test.syncCount, *retrieved.SyncCount)
			} else {
				assert.Nil(t, retrieved.SyncCount)
			}
			
			if test.syncDate != nil {
				assert.NotNil(t, retrieved.SyncDateFrom)
			} else {
				assert.Nil(t, retrieved.SyncDateFrom)
			}
		}
	})
}
