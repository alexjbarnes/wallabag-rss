// Package worker orchestrates fetching RSS feeds and sending articles to Wallabag on a scheduled basis.
package worker

import (
	"context"
	"fmt"
	"time"

	"wallabag-rss-tool/pkg/database"
	"wallabag-rss-tool/pkg/logging"
	"wallabag-rss-tool/pkg/models"
	"wallabag-rss-tool/pkg/rss"
	"wallabag-rss-tool/pkg/wallabag"
)

// Worker orchestrates fetching RSS feeds and sending articles to Wallabag.
type Worker struct {
	store          database.Storer
	rssProcessor   rss.Processorer
	wallabagClient wallabag.Clienter
	stopChan       chan struct{}
	priorityQueue  chan int // Channel for immediate feed processing
}

// NewWorker creates a new Worker instance.
func NewWorker(store database.Storer, rssProcessor rss.Processorer, wallabagClient wallabag.Clienter) *Worker {
	return &Worker{
		store:          store,
		rssProcessor:   rssProcessor,
		wallabagClient: wallabagClient,
		stopChan:       make(chan struct{}),
		priorityQueue:  make(chan int, 100), // Buffered channel to prevent blocking
	}
}

// Start begins the worker's polling loop.
func (w *Worker) Start() {
	logging.Info("Worker started")
	go w.run()
	go w.processPriorityQueue()
}

// Stop signals the worker to stop its polling loop.
func (w *Worker) Stop() {
	logging.Info("Worker stopping...")
	close(w.stopChan)
	// priorityQueue is left open to avoid panic if QueueFeedForImmediate is called during shutdown
}

func (w *Worker) run() {
	// Initial run immediately
	w.ProcessFeeds()

	// Get default poll interval from settings
	defaultInterval, err := w.store.GetDefaultPollInterval(context.Background())
	if err != nil {
		logging.Warn("Error getting default poll interval, using fallback",
			"error", err,
			"fallback_minutes", 60)
		defaultInterval = 60 // Fallback
	}

	logging.Info("Worker polling configured", "interval_minutes", defaultInterval)
	ticker := time.NewTicker(time.Duration(defaultInterval) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.ProcessFeeds()
		case <-w.stopChan:
			logging.Info("Worker stopped")

			return
		}
	}
}

// processPriorityQueue handles immediate feed processing requests
func (w *Worker) processPriorityQueue() {
	for {
		select {
		case feedID := <-w.priorityQueue:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			
			logging.Info("Processing priority feed from queue", "feed_id", feedID)
			
			if err := w.processSingleFeedByID(ctx, feedID); err != nil {
				logging.Error("Failed to process priority feed",
					"error", err,
					"feed_id", feedID)
			}
			
			cancel()
			
		case <-w.stopChan:
			logging.Info("Priority queue processor stopped")
			return
		}
	}
}

// QueueFeedForImmediate adds a feed to the priority queue for immediate processing
func (w *Worker) QueueFeedForImmediate(feedID int) {
	select {
	case w.priorityQueue <- feedID:
		logging.Info("Feed queued for immediate processing", "feed_id", feedID)
	default:
		// Channel is full, log warning but don't block
		logging.Warn("Priority queue is full, feed will be processed in next scheduled run",
			"feed_id", feedID,
			"queue_capacity", cap(w.priorityQueue))
	}
}

// QueueAllFeedsForImmediate queues all feeds for immediate processing (used for manual sync)
func (w *Worker) QueueAllFeedsForImmediate(ctx context.Context) error {
	feeds, err := w.store.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}
	
	queuedCount := 0
queueLoop:
	for _, feed := range feeds {
		select {
		case w.priorityQueue <- feed.ID:
			queuedCount++
		default:
			logging.Warn("Priority queue full, remaining feeds will sync on schedule",
				"queued", queuedCount,
				"total", len(feeds))

			break queueLoop
		}
	}
	
	logging.Info("Queued feeds for immediate processing",
		"queued_count", queuedCount,
		"total_count", len(feeds))

	return nil
}

// GetQueueStats returns statistics about the priority queue
func (w *Worker) GetQueueStats() (int, int) {
	return len(w.priorityQueue), cap(w.priorityQueue)
}

// ProcessFeeds fetches all active feeds and processes them.
func (w *Worker) ProcessFeeds() {
	w.ProcessFeedsWithContext(context.Background())
}

// ProcessFeedsWithContext fetches all active feeds and processes them with context support.
func (w *Worker) ProcessFeedsWithContext(ctx context.Context) {
	logging.Info("Processing feeds started")
	feeds, err := w.store.GetFeeds(ctx)
	if err != nil {
		logging.Error("Failed to get feeds from database", "error", fmt.Errorf("store.GetFeeds: %w", err))

		return
	}

	logging.Info("Retrieved feeds for processing", "feed_count", len(feeds))

	for _, feed := range feeds {
		if w.shouldStopProcessing(ctx) {
			return
		}

		w.processSingleFeed(ctx, &feed)
	}
	logging.Info("Processing feeds completed")
}

// processSingleFeedByID processes a single feed by its ID immediately
func (w *Worker) processSingleFeedByID(ctx context.Context, feedID int) error {
	feed, err := w.store.GetFeedByID(ctx, feedID)
	if err != nil {
		return fmt.Errorf("store.GetFeedByID: %w", err)
	}

	logging.Info("Processing single feed immediately",
		"feed_id", feed.ID,
		"feed_name", feed.Name,
		"feed_url", feed.URL)

	w.processSingleFeed(ctx, feed)

	return nil
}

// shouldStopProcessing checks if context is canceled
func (w *Worker) shouldStopProcessing(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		logging.Info("Feed processing canceled by context", "reason", ctx.Err())

		return true
	default:

		return false
	}
}

// processSingleFeed processes a single feed
func (w *Worker) processSingleFeed(ctx context.Context, feed *models.Feed) {
	feedLogger := logging.With("feed_id", feed.ID, "feed_name", feed.Name, "feed_url", feed.URL)

	// Check if it's time to fetch this feed
	effectiveInterval := w.getEffectiveInterval(ctx, feedLogger, feed)
	if w.shouldSkipFeed(feedLogger, feed, effectiveInterval) {
		return
	}

	// Fetch articles
	articles := w.fetchFeedArticles(feedLogger, feed)
	if articles == nil {
		return // Error already logged
	}

	// Process articles
	stats := w.processArticles(ctx, feedLogger, feed, articles)

	// Log results and update feed
	w.finalizeFeedProcessing(ctx, feedLogger, feed, articles, stats)
}

// getEffectiveInterval determines the effective polling interval for a feed
func (w *Worker) getEffectiveInterval(ctx context.Context, feedLogger logging.Logger, feed *models.Feed) int {
	effectiveInterval := feed.PollIntervalMinutes
	if effectiveInterval == 0 {
		defaultInterval, err := w.store.GetDefaultPollInterval(ctx)
		if err != nil {
			feedLogger.Warn("Error getting default poll interval, using fallback",
				"error", fmt.Errorf("store.GetDefaultPollInterval: %w", err),
				"fallback_minutes", 60)

			return 60
		}
		effectiveInterval = defaultInterval
	}

	return effectiveInterval
}

// shouldSkipFeed checks if a feed should be skipped based on timing
func (w *Worker) shouldSkipFeed(feedLogger logging.Logger, feed *models.Feed, effectiveInterval int) bool {
	if feed.LastFetched != nil && time.Since(*feed.LastFetched) < time.Duration(effectiveInterval)*time.Minute {
		nextFetch := time.Duration(effectiveInterval)*time.Minute - time.Since(*feed.LastFetched)
		feedLogger.Debug("Skipping feed, not yet time to fetch",
			"next_fetch_in", nextFetch.Round(time.Second),
			"poll_interval_minutes", effectiveInterval)

		return true
	}

	return false
}

// fetchFeedArticles fetches articles for a feed based on sync status
func (w *Worker) fetchFeedArticles(feedLogger logging.Logger, feed *models.Feed) []rss.Article {
	feedLogger.Info("Fetching articles for feed",
		"sync_mode", feed.SyncMode,
		"initial_sync_done", feed.InitialSyncDone)

	var articles []rss.Article
	var err error

	if !feed.InitialSyncDone {
		articles, err = w.rssProcessor.FetchAndParseWithSyncOptions(feed.URL, feed.SyncMode, feed.SyncCount, feed.SyncDateFrom)
		if err != nil {
			feedLogger.Error("Failed to fetch and parse feed for initial sync",
				"error", fmt.Errorf("rssProcessor.FetchAndParseWithSyncOptions: %w", err))

			return nil
		}
		feedLogger.Info("Initial sync completed",
			"articles_found", len(articles),
			"sync_mode", feed.SyncMode)
	} else {
		articles, err = w.rssProcessor.FetchAndParse(feed.URL)
		if err != nil {
			feedLogger.Error("Failed to fetch and parse feed",
				"error", fmt.Errorf("rssProcessor.FetchAndParse: %w", err))

			return nil
		}
		feedLogger.Debug("Regular sync completed", "articles_found", len(articles))
	}

	return articles
}

// ProcessingStats holds statistics for article processing
type ProcessingStats struct {
	ProcessedCount int
	NewCount       int
	ErrorCount     int
}

// processArticles processes all articles for a feed
func (w *Worker) processArticles(ctx context.Context, feedLogger logging.Logger, feed *models.Feed, articles []rss.Article) ProcessingStats {
	stats := ProcessingStats{}

	for _, article := range articles {
		if w.shouldStopProcessing(ctx) {
			feedLogger.Info("Article processing canceled by context", "reason", ctx.Err())

			return stats
		}

		w.processIndividualArticle(ctx, feedLogger, feed, article, &stats)
	}

	return stats
}

// processIndividualArticle processes a single article
func (w *Worker) processIndividualArticle(ctx context.Context, feedLogger logging.Logger, feed *models.Feed, article rss.Article, stats *ProcessingStats) {
	articleLogger := feedLogger.With("article_title", article.Title, "article_url", article.URL)

	processed, err := w.store.IsArticleAlreadyProcessed(ctx, article.URL)
	if err != nil {
		articleLogger.Error("Failed to check if article is already processed",
			"error", fmt.Errorf("store.IsArticleAlreadyProcessed: %w", err))
		stats.ErrorCount++

		return
	}
	if processed {
		articleLogger.Debug("Article already processed, skipping")
		stats.ProcessedCount++

		return
	}

	articleLogger.Info("Processing new article")
	wallabagEntry, err := w.wallabagClient.AddEntry(ctx, article.URL)
	if err != nil {
		articleLogger.Error("Failed to add article to Wallabag",
			"error", fmt.Errorf("wallabagClient.AddEntry: %w", err))
		stats.ErrorCount++

		return
	}

	articleLogger.Info("Article successfully added to Wallabag", "wallabag_entry_id", wallabagEntry.ID)

	// Convert and save article
	modelArticle := models.Article{
		Title:       article.Title,
		URL:         article.URL,
		PublishedAt: article.PublishedAt,
	}

	if err := w.store.SaveArticle(ctx, feed.ID, &modelArticle, wallabagEntry.ID); err != nil {
		articleLogger.Error("Failed to save article to database",
			"error", fmt.Errorf("store.SaveArticle: %w", err),
			"wallabag_entry_id", wallabagEntry.ID)
		stats.ErrorCount++
	} else {
		stats.NewCount++
	}
}

// finalizeFeedProcessing logs results and updates feed status
func (w *Worker) finalizeFeedProcessing(ctx context.Context, feedLogger logging.Logger, feed *models.Feed, articles []rss.Article, stats ProcessingStats) {
	feedLogger.Info("Feed processing completed",
		"total_articles", len(articles),
		"new_articles", stats.NewCount,
		"already_processed", stats.ProcessedCount,
		"errors", stats.ErrorCount)

	if err := w.store.UpdateFeedLastFetched(ctx, feed.ID); err != nil {
		feedLogger.Error("Failed to update feed last fetched time",
			"error", fmt.Errorf("store.UpdateFeedLastFetched: %w", err))
	}

	// Mark initial sync as completed if this was the first sync
	if !feed.InitialSyncDone {
		if err := w.store.MarkFeedInitialSyncCompleted(ctx, feed.ID); err != nil {
			feedLogger.Error("Failed to mark initial sync as completed",
				"error", fmt.Errorf("store.MarkFeedInitialSyncCompleted: %w", err))
		} else {
			feedLogger.Info("Initial sync marked as completed")
		}
	}
}
