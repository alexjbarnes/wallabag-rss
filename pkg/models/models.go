// Package models contains the data structures used throughout the wallabag-rss-tool application.
package models

import (
	"time"
)

// SyncMode represents the type of historical sync to perform for a feed
type SyncMode string

const (
	SyncModeNone     SyncMode = "none"      // Only sync from now onwards
	SyncModeAll      SyncMode = "all"       // Sync all available articles
	SyncModeCount    SyncMode = "count"     // Sync last N articles
	SyncModeDateFrom SyncMode = "date_from" // Sync articles from specific date
)

// TimeUnit represents the unit of time for intervals
type TimeUnit string

const (
	TimeUnitMinutes TimeUnit = "minutes"
	TimeUnitHours   TimeUnit = "hours"
	TimeUnitDays    TimeUnit = "days"
)

// Feed represents an RSS feed stored in the database.
type Feed struct {
	LastFetched         *time.Time // Use pointer for nullable DATETIME
	SyncDateFrom        *time.Time // Date to sync from (for SyncModeDateFrom)
	SyncCount           *int       // Number of articles to sync (for SyncModeCount)
	URL                 string
	Name                string
	SyncMode            SyncMode // How to handle historical articles on initial sync
	PollIntervalUnit    TimeUnit // Unit for poll interval (minutes, hours, days)
	ID                  int
	PollInterval        int  // Poll interval value
	PollIntervalMinutes int  // Legacy field for backward compatibility, computed from PollInterval and PollIntervalUnit
	InitialSyncDone     bool // Whether initial historical sync has been completed
}

// GetPollIntervalMinutes calculates the poll interval in minutes based on the interval and unit
func (f *Feed) GetPollIntervalMinutes() int {
	if f.PollInterval <= 0 {
		return 0
	}

	switch f.PollIntervalUnit {
	case TimeUnitMinutes:

		return f.PollInterval
	case TimeUnitHours:

		return f.PollInterval * 60
	case TimeUnitDays:

		return f.PollInterval * 60 * 24
	default:
		// Fallback to minutes if unit is unknown

		return f.PollInterval
	}
}

// SetPollInterval sets the poll interval with the specified value and unit
func (f *Feed) SetPollInterval(value int, unit TimeUnit) {
	f.PollInterval = value
	f.PollIntervalUnit = unit
	f.PollIntervalMinutes = f.GetPollIntervalMinutes()
}

// Article represents an article from an RSS feed, stored in the database.
type Article struct {
	PublishedAt     *time.Time
	WallabagEntryID *int // Use pointer for nullable INTEGER
	CreatedAt       time.Time
	Title           string
	URL             string
	ID              int
	FeedID          int
}
