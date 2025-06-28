package database_test

import (
	"context"
	"database/sql"
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
