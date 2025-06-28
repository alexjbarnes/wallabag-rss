// Package rss handles RSS feed fetching, parsing and processing with various sync options.
package rss

import (
	"fmt"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
	"wallabag-rss-tool/pkg/logging"
	"wallabag-rss-tool/pkg/models"
)

// Processorer defines the interface for RSS feed processing.
type Processorer interface {
	FetchAndParse(feedURL string) ([]Article, error)
	FetchAndParseWithSyncOptions(feedURL string, syncMode models.SyncMode, syncCount *int, syncDateFrom *time.Time) ([]Article, error)
}

// Article represents a simplified article structure from an RSS feed.
type Article struct {
	PublishedAt *time.Time
	Title       string
	URL         string
}

// Processor handles fetching and parsing RSS feeds.
type Processor struct {
	feedParser *gofeed.Parser
}

// NewProcessor creates a new RSS Processor.
func NewProcessor() *Processor {
	return &Processor{
		feedParser: gofeed.NewParser(),
	}
}

// FetchAndParse fetches an RSS feed from the given URL and parses it.
func (p *Processor) FetchAndParse(feedURL string) ([]Article, error) {
	logging.Debug("Fetching RSS feed", "feed_url", feedURL)
	feed, err := p.feedParser.ParseURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("feedParser.ParseURL failed for %s: %w", feedURL, err)
	}

	articles := make([]Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		if item.Link == "" || item.Title == "" {
			logging.Warn("Skipping RSS item with missing link or title",
				"feed_url", feedURL,
				"item_title", item.Title,
				"item_link", item.Link)

			continue
		}

		article := Article{
			Title: item.Title,
			URL:   item.Link,
		}
		if item.PublishedParsed != nil {
			article.PublishedAt = item.PublishedParsed
		} else if feed.PublishedParsed != nil {
			// Fallback to feed's published date if item's is missing
			article.PublishedAt = feed.PublishedParsed
		} else {
			// If no published date, use current time as a last resort
			now := time.Now()
			article.PublishedAt = &now
		}
		articles = append(articles, article)
	}

	logging.Info("Successfully fetched and parsed RSS feed",
		"feed_url", feedURL,
		"article_count", len(articles))

	return articles, nil
}

// FetchAndParseWithSyncOptions fetches and parses RSS feed with filtering based on sync options
func (p *Processor) FetchAndParseWithSyncOptions(feedURL string, syncMode models.SyncMode, syncCount *int, syncDateFrom *time.Time) ([]Article, error) {
	// First fetch all articles
	allArticles, err := p.FetchAndParse(feedURL)
	if err != nil {
		return nil, fmt.Errorf("FetchAndParse failed: %w", err)
	}

	// Apply filtering based on sync mode

	return p.applySyncFiltering(feedURL, allArticles, syncMode, syncCount, syncDateFrom)
}

// applySyncFiltering applies sync mode filtering to articles
func (p *Processor) applySyncFiltering(feedURL string, allArticles []Article, syncMode models.SyncMode, syncCount *int, syncDateFrom *time.Time) ([]Article, error) {
	switch syncMode {
	case models.SyncModeNone:

		return p.handleSyncModeNone(feedURL)
	case models.SyncModeAll:

		return p.handleSyncModeAll(feedURL, allArticles)
	case models.SyncModeCount:

		return p.handleSyncModeCount(feedURL, allArticles, syncCount)
	case models.SyncModeDateFrom:

		return p.handleSyncModeDateFrom(feedURL, allArticles, syncDateFrom)
	default:

		return p.handleUnknownSyncMode(feedURL, syncMode)
	}
}

// handleSyncModeNone handles the 'none' sync mode
func (p *Processor) handleSyncModeNone(feedURL string) ([]Article, error) {
	// Return no articles for historical sync (only future articles will be processed)
	logging.Debug("Sync mode 'none': returning no historical articles", "feed_url", feedURL)

	return []Article{}, nil
}

// handleSyncModeAll handles the 'all' sync mode
func (p *Processor) handleSyncModeAll(feedURL string, allArticles []Article) ([]Article, error) {
	// Sort all articles oldest first for chronological processing
	sortedArticles := p.sortArticlesByDate(allArticles)
	
	logging.Debug("Sync mode 'all': returning all articles in chronological order",
		"feed_url", feedURL,
		"article_count", len(sortedArticles))

	return sortedArticles, nil
}

// handleSyncModeCount handles the 'count' sync mode
func (p *Processor) handleSyncModeCount(feedURL string, allArticles []Article, syncCount *int) ([]Article, error) {
	if syncCount == nil || *syncCount <= 0 {
		logging.Warn("Sync mode 'count' but no valid count provided",
			"feed_url", feedURL,
			"sync_count", syncCount)

		return []Article{}, nil
	}

	// Sort articles by published date (newest first) to get the most recent N
	sortedNewestFirst := p.sortArticlesByDateNewestFirst(allArticles)

	// Get the most recent N articles
	count := *syncCount
	if count > len(sortedNewestFirst) {
		count = len(sortedNewestFirst)
	}
	
	recentArticles := sortedNewestFirst[:count]
	
	// Now sort these N articles oldest first for processing
	finalArticles := p.sortArticlesByDate(recentArticles)
	
	logging.Debug("Sync mode 'count': returning most recent articles in chronological order",
		"feed_url", feedURL,
		"returned_count", count,
		"total_articles", len(allArticles))

	return finalArticles, nil
}

// handleSyncModeDateFrom handles the 'date_from' sync mode
func (p *Processor) handleSyncModeDateFrom(feedURL string, allArticles []Article, syncDateFrom *time.Time) ([]Article, error) {
	if syncDateFrom == nil {
		logging.Warn("Sync mode 'date_from' but no date provided", "feed_url", feedURL)

		return []Article{}, nil
	}

	// Filter articles published on or after the specified date
	filteredArticles := p.filterArticlesByDate(allArticles, syncDateFrom)
	
	// Sort filtered articles oldest first for chronological processing
	sortedArticles := p.sortArticlesByDate(filteredArticles)
	
	logging.Debug("Sync mode 'date_from': returning articles after date in chronological order",
		"feed_url", feedURL,
		"filtered_count", len(sortedArticles),
		"sync_date_from", p.formatDateOrNil(syncDateFrom),
		"total_articles", len(allArticles))

	return sortedArticles, nil
}

// handleUnknownSyncMode handles unknown sync modes
func (p *Processor) handleUnknownSyncMode(feedURL string, syncMode models.SyncMode) ([]Article, error) {
	logging.Warn("Unknown sync mode, treating as 'none'",
		"feed_url", feedURL,
		"sync_mode", syncMode)

	return []Article{}, nil
}

// sortArticlesByDate sorts articles by published date (oldest first)
func (p *Processor) sortArticlesByDate(articles []Article) []Article {
	sortedArticles := make([]Article, len(articles))
	copy(sortedArticles, articles)
	sort.Slice(sortedArticles, func(firstIdx, secondIdx int) bool {
		firstTime := sortedArticles[firstIdx].PublishedAt
		secondTime := sortedArticles[secondIdx].PublishedAt

		if firstTime == nil && secondTime == nil {
			return false
		}
		if firstTime == nil {
			return false
		}
		if secondTime == nil {
			return true
		}

		return firstTime.Before(*secondTime)
	})

	return sortedArticles
}

// sortArticlesByDateNewestFirst sorts articles by published date (newest first)
func (p *Processor) sortArticlesByDateNewestFirst(articles []Article) []Article {
	sortedArticles := make([]Article, len(articles))
	copy(sortedArticles, articles)
	sort.Slice(sortedArticles, func(firstIdx, secondIdx int) bool {
		firstTime := sortedArticles[firstIdx].PublishedAt
		secondTime := sortedArticles[secondIdx].PublishedAt

		if firstTime == nil && secondTime == nil {
			return false
		}
		if firstTime == nil {
			return false
		}
		if secondTime == nil {
			return true
		}

		return firstTime.After(*secondTime)
	})

	return sortedArticles
}

// filterArticlesByDate filters articles published on or after the specified date
func (p *Processor) filterArticlesByDate(articles []Article, syncDateFrom *time.Time) []Article {
	var filteredArticles []Article
	for _, article := range articles {
		if article.PublishedAt != nil && syncDateFrom != nil &&
			(article.PublishedAt.After(*syncDateFrom) || article.PublishedAt.Equal(*syncDateFrom)) {
			filteredArticles = append(filteredArticles, article)
		}
	}

	return filteredArticles
}

// formatDateOrNil formats a date pointer or returns "nil" if nil
func (p *Processor) formatDateOrNil(date *time.Time) string {
	if date != nil {
		return date.Format("2006-01-02")
	}

	return "nil"
}
