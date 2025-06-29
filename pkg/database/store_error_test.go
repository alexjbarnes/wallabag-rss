package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"wallabag-rss-tool/pkg/database"
	"wallabag-rss-tool/pkg/models"
)

func TestSQLStore_ErrorPaths(t *testing.T) {
	t.Run("UpdateFeed statement preparation error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		// Mock statement preparation failure
		mock.ExpectPrepare("UPDATE feeds SET").WillReturnError(errors.New("prepare failed"))

		feed := &models.Feed{
			ID:   1,
			Name: "Test Feed",
			URL:  "https://example.com/test.xml",
		}

		err = store.UpdateFeed(ctx, feed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare update feed statement")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateFeed statement execution error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		feed := &models.Feed{
			ID:   1,
			Name: "Test Feed",
			URL:  "https://example.com/test.xml",
		}
		feed.SetPollInterval(1, models.TimeUnitHours)

		// Mock successful preparation but failed execution
		mock.ExpectPrepare("UPDATE feeds SET").ExpectExec().
			WithArgs(feed.Name, feed.URL, feed.PollIntervalMinutes, feed.PollInterval, 
				string(feed.PollIntervalUnit), string(feed.SyncMode), nil, nil, feed.InitialSyncDone, feed.ID).
			WillReturnError(errors.New("execution failed"))

		err = store.UpdateFeed(ctx, feed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update feed")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteFeed statement preparation error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectPrepare("DELETE FROM feeds WHERE id = ?").WillReturnError(errors.New("prepare failed"))

		err = store.DeleteFeed(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare delete feed statement")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteFeed statement execution error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectPrepare("DELETE FROM feeds WHERE id = ?").ExpectExec().
			WithArgs(1).WillReturnError(errors.New("execution failed"))

		err = store.DeleteFeed(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete feed")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("InsertFeed statement preparation error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectPrepare("INSERT INTO feeds").WillReturnError(errors.New("prepare failed"))

		feed := &models.Feed{
			Name: "Test Feed",
			URL:  "https://example.com/test.xml",
		}

		_, err = store.InsertFeed(ctx, feed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare insert feed statement")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("InsertFeed statement execution error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		feed := &models.Feed{
			Name: "Test Feed",
			URL:  "https://example.com/test.xml",
		}
		feed.SetPollInterval(1, models.TimeUnitHours)

		mock.ExpectPrepare("INSERT INTO feeds").ExpectExec().
			WithArgs(feed.Name, feed.URL, feed.PollIntervalMinutes, feed.PollInterval,
				string(feed.PollIntervalUnit), string(feed.SyncMode), nil, nil, feed.InitialSyncDone).
			WillReturnError(errors.New("execution failed"))

		_, err = store.InsertFeed(ctx, feed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to insert feed")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("InsertFeed LastInsertId error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		feed := &models.Feed{
			Name: "Test Feed",
			URL:  "https://example.com/test.xml",
		}
		feed.SetPollInterval(1, models.TimeUnitHours)

		result := sqlmock.NewErrorResult(errors.New("last insert id failed"))
		mock.ExpectPrepare("INSERT INTO feeds").ExpectExec().
			WithArgs(feed.Name, feed.URL, feed.PollIntervalMinutes, feed.PollInterval,
				string(feed.PollIntervalUnit), string(feed.SyncMode), nil, nil, feed.InitialSyncDone).
			WillReturnResult(result)

		_, err = store.InsertFeed(ctx, feed)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get last insert ID")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateDefaultPollInterval statement preparation error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectPrepare("INSERT OR REPLACE INTO settings").WillReturnError(errors.New("prepare failed"))

		err = store.UpdateDefaultPollInterval(ctx, 60)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare update settings statement")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateDefaultPollInterval statement execution error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectPrepare("INSERT OR REPLACE INTO settings").ExpectExec().
			WithArgs("default_poll_interval_minutes", 60).
			WillReturnError(errors.New("execution failed"))

		err = store.UpdateDefaultPollInterval(ctx, 60)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update settings")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateFeedLastFetched statement preparation error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectPrepare("UPDATE feeds SET last_fetched = ?").WillReturnError(errors.New("prepare failed"))

		err = store.UpdateFeedLastFetched(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare update feed statement")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateFeedLastFetched statement execution error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectPrepare("UPDATE feeds SET last_fetched = ?").ExpectExec().
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnError(errors.New("execution failed"))

		err = store.UpdateFeedLastFetched(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update feed last_fetched")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("MarkFeedInitialSyncCompleted statement preparation error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectPrepare("UPDATE feeds SET initial_sync_done = 1").WillReturnError(errors.New("prepare failed"))

		err = store.MarkFeedInitialSyncCompleted(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare update feed sync statement")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("MarkFeedInitialSyncCompleted statement execution error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectPrepare("UPDATE feeds SET initial_sync_done = 1").ExpectExec().
			WithArgs(1).WillReturnError(errors.New("execution failed"))

		err = store.MarkFeedInitialSyncCompleted(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to mark feed initial sync completed")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveArticle statement preparation error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectPrepare("INSERT INTO articles").WillReturnError(errors.New("prepare failed"))

		article := &models.Article{
			Title: "Test Article",
			URL:   "https://example.com/article",
		}

		err = store.SaveArticle(ctx, 1, article, 123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare insert article statement")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveArticle statement execution error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		article := &models.Article{
			Title: "Test Article",
			URL:   "https://example.com/article",
		}

		mock.ExpectPrepare("INSERT INTO articles").ExpectExec().
			WithArgs(1, article.Title, article.URL, 123, article.PublishedAt).
			WillReturnError(errors.New("execution failed"))

		err = store.SaveArticle(ctx, 1, article, 123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to insert article")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetFeeds query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectQuery("SELECT").WillReturnError(errors.New("query failed"))

		feeds, err := store.GetFeeds(ctx)
		assert.Error(t, err)
		assert.Nil(t, feeds)
		assert.Contains(t, err.Error(), "failed to query feeds")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetFeeds rows iteration error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		rows := sqlmock.NewRows([]string{"id", "url", "name", "last_fetched", "poll_interval", "poll_interval_unit", "sync_mode", "sync_count", "sync_date_from", "initial_sync_done"}).
			AddRow(1, "https://example.com", "Test", nil, 1, "hours", "none", nil, nil, false).
			RowError(0, errors.New("row error"))

		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		feeds, err := store.GetFeeds(ctx)
		assert.Error(t, err)
		assert.Nil(t, feeds)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetArticles query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectQuery("SELECT id, feed_id, title, url").WillReturnError(errors.New("query failed"))

		articles, err := store.GetArticles(ctx)
		assert.Error(t, err)
		assert.Nil(t, articles)
		assert.Contains(t, err.Error(), "failed to query articles")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetArticles rows iteration error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		rows := sqlmock.NewRows([]string{"id", "feed_id", "title", "url", "wallabag_entry_id", "published_at", "created_at"}).
			AddRow(1, 1, "Test Article", "https://example.com", nil, nil, time.Now()).
			RowError(0, errors.New("row error"))

		mock.ExpectQuery("SELECT id, feed_id, title, url").WillReturnRows(rows)

		articles, err := store.GetArticles(ctx)
		assert.Error(t, err)
		assert.Nil(t, articles)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("IsArticleAlreadyProcessed query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		store := database.NewSQLStore(db)
		ctx := context.Background()

		mock.ExpectQuery("SELECT COUNT").WillReturnError(errors.New("query failed"))

		processed, err := store.IsArticleAlreadyProcessed(ctx, "https://example.com")
		assert.Error(t, err)
		assert.False(t, processed)
		assert.Contains(t, err.Error(), "error checking for existing article")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}