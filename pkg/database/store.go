package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"wallabag-rss-tool/pkg/logging"
	"wallabag-rss-tool/pkg/models"
)

// Storer defines the interface for database operations.
type Storer interface {
	GetFeeds(ctx context.Context) ([]models.Feed, error)
	GetFeedByID(ctx context.Context, id int) (*models.Feed, error)
	InsertFeed(ctx context.Context, feed *models.Feed) (int64, error)
	UpdateFeed(ctx context.Context, feed *models.Feed) error
	DeleteFeed(ctx context.Context, id int) error
	GetArticles(ctx context.Context) ([]models.Article, error)
	SaveArticle(ctx context.Context, feedID int, article *models.Article, wallabagEntryID int) error
	IsArticleAlreadyProcessed(ctx context.Context, articleURL string) (bool, error)
	GetDefaultPollInterval(ctx context.Context) (int, error)
	UpdateDefaultPollInterval(ctx context.Context, interval int) error
	UpdateFeedLastFetched(ctx context.Context, feedID int) error
	MarkFeedInitialSyncCompleted(ctx context.Context, feedID int) error
}

// SQLStore implements Storer using a SQL database.
type SQLStore struct {
	db *sql.DB
}

// NewSQLStore creates a new SQLStore.
func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

// GetFeeds retrieves all feeds from the database.
func (s *SQLStore) GetFeeds(ctx context.Context) ([]models.Feed, error) {
	query := `
		SELECT 
			id, url, name, last_fetched,
			COALESCE(poll_interval, 1) as poll_interval,
			COALESCE(poll_interval_unit, 'days') as poll_interval_unit,
			sync_mode, sync_count, sync_date_from, initial_sync_done 
		FROM feeds
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query feeds: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logging.Error("Failed to close feed rows", "error", err)
		}
	}()

	feeds := make([]models.Feed, 0, 10) // Pre-allocate with reasonable capacity
	for rows.Next() {
		feed, err := s.scanFeedRow(rows)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over feed rows: %w", err)
	}

	return feeds, nil
}

// scanFeedRow scans a single feed row from the database
func (s *SQLStore) scanFeedRow(rows *sql.Rows) (models.Feed, error) {
	var feed models.Feed
	var lastFetched sql.NullTime
	var pollInterval sql.NullInt64
	var pollIntervalUnit sql.NullString
	var syncMode sql.NullString
	var syncCount sql.NullInt64
	var syncDateFrom sql.NullTime
	var initialSyncDone sql.NullBool

	if err := rows.Scan(&feed.ID, &feed.URL, &feed.Name, &lastFetched,
		&pollInterval, &pollIntervalUnit, &syncMode, &syncCount, &syncDateFrom, &initialSyncDone); err != nil {
		return models.Feed{}, fmt.Errorf("failed to scan feed row: %w", err)
	}

	s.setFeedNullableFields(&feed, lastFetched, pollInterval, pollIntervalUnit, syncMode, syncCount, syncDateFrom, initialSyncDone)

	return feed, nil
}

// setFeedNullableFields sets nullable database fields on the feed model
func (s *SQLStore) setFeedNullableFields(feed *models.Feed, lastFetched sql.NullTime, pollInterval sql.NullInt64, pollIntervalUnit, syncMode sql.NullString, syncCount sql.NullInt64, syncDateFrom sql.NullTime, initialSyncDone sql.NullBool) {
	if lastFetched.Valid {
		feed.LastFetched = &lastFetched.Time
	}

	if pollInterval.Valid {
		feed.PollInterval = int(pollInterval.Int64)
	} else {
		feed.PollInterval = 1 // Default
	}

	if pollIntervalUnit.Valid {
		feed.PollIntervalUnit = models.TimeUnit(pollIntervalUnit.String)
	} else {
		feed.PollIntervalUnit = models.TimeUnitDays // Default
	}

	// Compute poll interval in minutes
	feed.PollIntervalMinutes = feed.GetPollIntervalMinutes()

	if syncMode.Valid {
		feed.SyncMode = models.SyncMode(syncMode.String)
	} else {
		feed.SyncMode = models.SyncModeNone
	}

	if syncCount.Valid {
		count := int(syncCount.Int64)
		feed.SyncCount = &count
	}

	if syncDateFrom.Valid {
		feed.SyncDateFrom = &syncDateFrom.Time
	}

	if initialSyncDone.Valid {
		feed.InitialSyncDone = initialSyncDone.Bool
	}
}

// GetFeedByID retrieves a single feed by its ID.
func (s *SQLStore) GetFeedByID(ctx context.Context, id int) (*models.Feed, error) {
	var feed models.Feed
	var lastFetched sql.NullTime
	var pollInterval sql.NullInt64
	var pollIntervalUnit sql.NullString
	var syncMode sql.NullString
	var syncCount sql.NullInt64
	var syncDateFrom sql.NullTime
	var initialSyncDone sql.NullBool

	query := `
		SELECT 
			id, url, name, last_fetched,
			COALESCE(poll_interval, 1) as poll_interval,
			COALESCE(poll_interval_unit, 'days') as poll_interval_unit,
			sync_mode, sync_count, sync_date_from, initial_sync_done 
		FROM feeds WHERE id = ?
	`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&feed.ID, &feed.URL, &feed.Name, &lastFetched,
		&pollInterval, &pollIntervalUnit, &syncMode, &syncCount, &syncDateFrom, &initialSyncDone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("feed with ID %d not found", id)
		}

		return nil, fmt.Errorf("failed to query feed by ID: %w", err)
	}

	s.setFeedNullableFields(&feed, lastFetched, pollInterval, pollIntervalUnit, syncMode, syncCount, syncDateFrom, initialSyncDone)

	return &feed, nil
}

// InsertFeed inserts a new feed into the database.
func (s *SQLStore) InsertFeed(ctx context.Context, feed *models.Feed) (int64, error) {
	stmt, err := s.db.PrepareContext(ctx, `
		INSERT INTO feeds (
			name, url, poll_interval_minutes, poll_interval, poll_interval_unit, 
			sync_mode, sync_count, sync_date_from, initial_sync_done
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare insert feed statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logging.Error("Failed to close statement", "error", err)
		}
	}()

	var syncCount interface{}
	if feed.SyncCount != nil {
		syncCount = *feed.SyncCount
	}

	var syncDateFrom interface{}
	if feed.SyncDateFrom != nil {
		syncDateFrom = *feed.SyncDateFrom
	}

	// Ensure PollIntervalMinutes is calculated
	feed.PollIntervalMinutes = feed.GetPollIntervalMinutes()

	res, err := stmt.Exec(
		feed.Name, feed.URL, feed.PollIntervalMinutes,
		feed.PollInterval, string(feed.PollIntervalUnit),
		string(feed.SyncMode), syncCount, syncDateFrom, feed.InitialSyncDone)
	if err != nil {
		return 0, fmt.Errorf("failed to insert feed: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// UpdateFeed updates an existing feed in the database.
func (s *SQLStore) UpdateFeed(ctx context.Context, feed *models.Feed) error {
	stmt, err := s.db.PrepareContext(ctx, `
		UPDATE feeds SET 
			name = ?, url = ?, poll_interval_minutes = ?, poll_interval = ?, poll_interval_unit = ?,
			sync_mode = ?, sync_count = ?, sync_date_from = ?, initial_sync_done = ? 
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare update feed statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logging.Error("Failed to close statement", "error", err)
		}
	}()

	var syncCount interface{}
	if feed.SyncCount != nil {
		syncCount = *feed.SyncCount
	}

	var syncDateFrom interface{}
	if feed.SyncDateFrom != nil {
		syncDateFrom = *feed.SyncDateFrom
	}

	// Ensure PollIntervalMinutes is calculated
	feed.PollIntervalMinutes = feed.GetPollIntervalMinutes()

	_, err = stmt.Exec(
		feed.Name, feed.URL, feed.PollIntervalMinutes,
		feed.PollInterval, string(feed.PollIntervalUnit),
		string(feed.SyncMode), syncCount, syncDateFrom, feed.InitialSyncDone, feed.ID)
	if err != nil {
		return fmt.Errorf("failed to update feed: %w", err)
	}

	return nil
}

// DeleteFeed deletes a feed from the database.
func (s *SQLStore) DeleteFeed(ctx context.Context, id int) error {
	stmt, err := s.db.PrepareContext(ctx, "DELETE FROM feeds WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare delete feed statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logging.Error("Failed to close statement", "error", err)
		}
	}()

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to delete feed: %w", err)
	}

	return nil
}

// GetArticles retrieves all articles from the database.
func (s *SQLStore) GetArticles(ctx context.Context) ([]models.Article, error) {
	rows, err := s.db.Query("SELECT id, feed_id, title, url, wallabag_entry_id, published_at, created_at FROM articles ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query articles: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logging.Error("Failed to close article rows", "error", err)
		}
	}()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		var wallabagEntryID sql.NullInt64
		var publishedAt sql.NullTime

		if err := rows.Scan(&article.ID, &article.FeedID, &article.Title, &article.URL, &wallabagEntryID, &publishedAt, &article.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan article row: %w", err)
		}
		if wallabagEntryID.Valid {
			id := int(wallabagEntryID.Int64)
			article.WallabagEntryID = &id
		}
		if publishedAt.Valid {
			article.PublishedAt = &publishedAt.Time
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over article rows: %w", err)
	}

	return articles, nil
}

// SaveArticle saves a new article to the database.
func (s *SQLStore) SaveArticle(ctx context.Context, feedID int, article *models.Article, wallabagEntryID int) error {
	stmt, err := s.db.PrepareContext(ctx,
		"INSERT INTO articles (feed_id, title, url, wallabag_entry_id, published_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare insert article statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logging.Error("Failed to close statement", "error", err)
		}
	}()

	_, err = stmt.Exec(feedID, article.Title, article.URL, wallabagEntryID, article.PublishedAt)
	if err != nil {
		return fmt.Errorf("failed to insert article: %w", err)
	}

	return nil
}

// IsArticleAlreadyProcessed checks if an article with the given URL already exists in the database.
func (s *SQLStore) IsArticleAlreadyProcessed(ctx context.Context, articleURL string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM articles WHERE url = ?", articleURL).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking for existing article: %w", err)
	}

	return count > 0, nil
}

// GetDefaultPollInterval retrieves the default poll interval from settings.
func (s *SQLStore) GetDefaultPollInterval(ctx context.Context) (int, error) {
	var interval int
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", "default_poll_interval_minutes").Scan(&interval)
	if err != nil {
		return 0, fmt.Errorf("failed to get default poll interval from settings: %w", err)
	}

	return interval, nil
}

// UpdateDefaultPollInterval updates the default poll interval in settings.
func (s *SQLStore) UpdateDefaultPollInterval(ctx context.Context, interval int) error {
	stmt, err := s.db.PrepareContext(ctx, "INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare update settings statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logging.Error("Failed to close statement", "error", err)
		}
	}()

	_, err = stmt.Exec("default_poll_interval_minutes", interval)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}

// UpdateFeedLastFetched updates the last_fetched timestamp for a feed.
func (s *SQLStore) UpdateFeedLastFetched(ctx context.Context, feedID int) error {
	stmt, err := s.db.PrepareContext(ctx, "UPDATE feeds SET last_fetched = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare update feed statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logging.Error("Failed to close statement", "error", err)
		}
	}()

	_, err = stmt.Exec(time.Now(), feedID)
	if err != nil {
		return fmt.Errorf("failed to update feed last_fetched: %w", err)
	}

	return nil
}

// MarkFeedInitialSyncCompleted marks a feed's initial sync as completed.
func (s *SQLStore) MarkFeedInitialSyncCompleted(ctx context.Context, feedID int) error {
	stmt, err := s.db.PrepareContext(ctx, "UPDATE feeds SET initial_sync_done = 1 WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare update feed sync statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			logging.Error("Failed to close statement", "error", err)
		}
	}()

	_, err = stmt.Exec(feedID)
	if err != nil {
		return fmt.Errorf("failed to mark feed initial sync completed: %w", err)
	}

	return nil
}
