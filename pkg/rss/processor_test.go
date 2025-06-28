package rss_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"wallabag-rss-tool/pkg/models"
	"wallabag-rss-tool/pkg/rss"
)

func TestNewProcessor(t *testing.T) {
	processor := rss.NewProcessor()
	assert.NotNil(t, processor)
	// Cannot access unexported field feedParser from external test
}

func TestArticleStruct(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		article rss.Article
	}{
		{
			name: "Complete article",
			article: rss.Article{
				Title:       "Test Article",
				URL:         "https://example.com/article1",
				PublishedAt: &now,
			},
		},
		{
			name: "Article without published date",
			article: rss.Article{
				Title:       "Another Article",
				URL:         "https://example.com/article2",
				PublishedAt: nil,
			},
		},
		{
			name:    "Zero values",
			article: rss.Article{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := tt.article

			assert.Equal(t, tt.article.Title, article.Title)
			assert.Equal(t, tt.article.URL, article.URL)
			assert.Equal(t, tt.article.PublishedAt, article.PublishedAt)
		})
	}
}

func TestProcessor_FetchAndParse(t *testing.T) {
	processor := rss.NewProcessor()

	t.Run("Valid RSS feed", func(t *testing.T) {
		// Create a test server with valid RSS
		validRSS := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Feed</title>
		<description>Test RSS Feed</description>
		<link>https://example.com</link>
		<pubDate>Mon, 01 Jan 2024 12:00:00 GMT</pubDate>
		<item>
			<title>First Article</title>
			<link>https://example.com/article1</link>
			<description>Description of first article</description>
			<pubDate>Mon, 01 Jan 2024 10:00:00 GMT</pubDate>
		</item>
		<item>
			<title>Second Article</title>
			<link>https://example.com/article2</link>
			<description>Description of second article</description>
			<pubDate>Mon, 01 Jan 2024 11:00:00 GMT</pubDate>
		</item>
	</channel>
</rss>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(validRSS))
		}))
		defer server.Close()

		articles, err := processor.FetchAndParse(server.URL)
		assert.NoError(t, err)
		assert.Len(t, articles, 2)

		// Check first article
		article1 := articles[0]
		assert.Equal(t, "First Article", article1.Title)
		assert.Equal(t, "https://example.com/article1", article1.URL)
		assert.NotNil(t, article1.PublishedAt)

		// Check second article
		article2 := articles[1]
		assert.Equal(t, "Second Article", article2.Title)
		assert.Equal(t, "https://example.com/article2", article2.URL)
		assert.NotNil(t, article2.PublishedAt)
	})

	t.Run("RSS feed with missing item fields", func(t *testing.T) {
		// RSS with some items missing title or link
		incompleteRSS := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Feed</title>
		<description>Test RSS Feed</description>
		<link>https://example.com</link>
		<item>
			<title>Complete Article</title>
			<link>https://example.com/complete</link>
			<description>Complete article</description>
		</item>
		<item>
			<title></title>
			<link>https://example.com/empty-title</link>
			<description>Article with empty title</description>
		</item>
		<item>
			<title>Missing Link Article</title>
			<link></link>
			<description>Article with empty link</description>
		</item>
		<item>
			<description>Article with no title or link</description>
		</item>
	</channel>
</rss>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(incompleteRSS))
		}))
		defer server.Close()

		articles, err := processor.FetchAndParse(server.URL)
		assert.NoError(t, err)
		// Should only return the complete article, skipping incomplete ones
		assert.Len(t, articles, 1)
		assert.Equal(t, "Complete Article", articles[0].Title)
		assert.Equal(t, "https://example.com/complete", articles[0].URL)
	})

	t.Run("RSS feed with missing published dates", func(t *testing.T) {
		// RSS with no item published dates, but feed-level published date
		noPubDateRSS := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Feed</title>
		<description>Test RSS Feed</description>
		<link>https://example.com</link>
		<pubDate>Mon, 01 Jan 2024 12:00:00 GMT</pubDate>
		<item>
			<title>Article Without PubDate</title>
			<link>https://example.com/no-pubdate</link>
			<description>Article without publication date</description>
		</item>
	</channel>
</rss>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(noPubDateRSS))
		}))
		defer server.Close()

		articles, err := processor.FetchAndParse(server.URL)
		assert.NoError(t, err)
		assert.Len(t, articles, 1)

		article := articles[0]
		assert.Equal(t, "Article Without PubDate", article.Title)
		assert.Equal(t, "https://example.com/no-pubdate", article.URL)
		// Should fallback to feed's published date
		assert.NotNil(t, article.PublishedAt)
	})

	t.Run("RSS feed with no published dates at all", func(t *testing.T) {
		// RSS with no published dates anywhere
		noDateRSS := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Feed</title>
		<description>Test RSS Feed</description>
		<link>https://example.com</link>
		<item>
			<title>Article Without Any Date</title>
			<link>https://example.com/no-date</link>
			<description>Article without any publication date</description>
		</item>
	</channel>
</rss>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(noDateRSS))
		}))
		defer server.Close()

		beforeParse := time.Now()
		articles, err := processor.FetchAndParse(server.URL)
		afterParse := time.Now()

		assert.NoError(t, err)
		assert.Len(t, articles, 1)

		article := articles[0]
		assert.Equal(t, "Article Without Any Date", article.Title)
		assert.Equal(t, "https://example.com/no-date", article.URL)
		// Should use current time as fallback
		assert.NotNil(t, article.PublishedAt)
		assert.True(t, article.PublishedAt.After(beforeParse) || article.PublishedAt.Equal(beforeParse))
		assert.True(t, article.PublishedAt.Before(afterParse) || article.PublishedAt.Equal(afterParse))
	})

	t.Run("Invalid URL", func(t *testing.T) {
		articles, err := processor.FetchAndParse("invalid-url")
		assert.Error(t, err)
		assert.Nil(t, articles)
		assert.Contains(t, err.Error(), "feedParser.ParseURL failed for invalid-url")
	})

	t.Run("URL not found", func(t *testing.T) {
		articles, err := processor.FetchAndParse("https://nonexistent.example.com/feed.rss")
		assert.Error(t, err)
		assert.Nil(t, articles)
		assert.Contains(t, err.Error(), "feedParser.ParseURL failed for https://nonexistent.example.com/feed.rss")
	})

	t.Run("Invalid RSS content", func(t *testing.T) {
		// Server returns invalid XML
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Not valid XML content"))
		}))
		defer server.Close()

		articles, err := processor.FetchAndParse(server.URL)
		assert.Error(t, err)
		assert.Nil(t, articles)
		assert.Contains(t, err.Error(), "feedParser.ParseURL failed for")
	})

	t.Run("Server error", func(t *testing.T) {
		// Server returns 500 error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		articles, err := processor.FetchAndParse(server.URL)
		assert.Error(t, err)
		assert.Nil(t, articles)
		assert.Contains(t, err.Error(), "feedParser.ParseURL failed for")
	})

	t.Run("Empty RSS feed", func(t *testing.T) {
		// Valid RSS but no items
		emptyRSS := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Empty Feed</title>
		<description>RSS Feed with no items</description>
		<link>https://example.com</link>
	</channel>
</rss>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(emptyRSS))
		}))
		defer server.Close()

		articles, err := processor.FetchAndParse(server.URL)
		assert.NoError(t, err)
		assert.Empty(t, articles)
	})

	t.Run("Atom feed", func(t *testing.T) {
		// Test that processor can handle Atom feeds too (gofeed supports both)
		atomFeed := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
	<title>Test Atom Feed</title>
	<link href="https://example.com"/>
	<updated>2024-01-01T12:00:00Z</updated>
	<id>https://example.com</id>
	<entry>
		<title>Atom Article</title>
		<link href="https://example.com/atom-article"/>
		<id>https://example.com/atom-article</id>
		<updated>2024-01-01T10:00:00Z</updated>
		<summary>Atom article description</summary>
	</entry>
</feed>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/atom+xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(atomFeed))
		}))
		defer server.Close()

		articles, err := processor.FetchAndParse(server.URL)
		assert.NoError(t, err)
		assert.Len(t, articles, 1)

		article := articles[0]
		assert.Equal(t, "Atom Article", article.Title)
		assert.Equal(t, "https://example.com/atom-article", article.URL)
		assert.NotNil(t, article.PublishedAt)
	})
}

func TestProcessor_Interface(t *testing.T) {
	t.Run("Processor implements Processorer interface", func(t *testing.T) {
		var processor rss.Processorer = rss.NewProcessor()
		assert.NotNil(t, processor)

		// Test that we can call interface methods
		assert.NotPanics(t, func() {
			// This will fail because it's an invalid URL, but it should not panic
			processor.FetchAndParse("invalid-url")
		})
	})
}

func TestProcessorer_Interface(t *testing.T) {
	t.Run("Interface method signature", func(t *testing.T) {
		// This test ensures the interface is correctly defined
		var processor rss.Processorer

		// Verify interface method exists and has correct signature
		assert.NotPanics(t, func() {
			if processor != nil {
				processor.FetchAndParse("test-url")
			}
		})
	})
}

func TestArticle_PointerFields(t *testing.T) {
	t.Run("PublishedAt pointer behavior", func(t *testing.T) {
		now := time.Now()

		// Test nil pointer
		article1 := rss.Article{PublishedAt: nil}
		assert.Nil(t, article1.PublishedAt)

		// Test non-nil pointer
		article2 := rss.Article{PublishedAt: &now}
		assert.NotNil(t, article2.PublishedAt)
		assert.Equal(t, now, *article2.PublishedAt)

		// Test pointer assignment
		article3 := rss.Article{}
		article3.PublishedAt = &now
		assert.NotNil(t, article3.PublishedAt)
		assert.Equal(t, now, *article3.PublishedAt)
	})
}

func TestArticle_Copy(t *testing.T) {
	t.Run("Article copy with pointer", func(t *testing.T) {
		now := time.Now()
		original := rss.Article{
			Title:       "Original Article",
			URL:         "https://example.com/original",
			PublishedAt: &now,
		}

		copied := original
		assert.Equal(t, original.Title, copied.Title)
		assert.Equal(t, original.URL, copied.URL)
		assert.Equal(t, original.PublishedAt, copied.PublishedAt)

		// Pointer should point to same time value
		if original.PublishedAt != nil && copied.PublishedAt != nil {
			assert.Equal(t, *original.PublishedAt, *copied.PublishedAt)
		}
	})
}

// Benchmark tests
func TestProcessor_FetchAndParseWithSyncOptions(t *testing.T) {
	processor := rss.NewProcessor()

	// Create a test server with RSS content that has known dates
	validRSS := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Feed</title>
		<description>Test RSS Feed</description>
		<link>https://example.com</link>
		<item>
			<title>Article from Jan 1</title>
			<link>https://example.com/article1</link>
			<description>Description 1</description>
			<pubDate>Mon, 01 Jan 2024 10:00:00 GMT</pubDate>
		</item>
		<item>
			<title>Article from Jan 2</title>
			<link>https://example.com/article2</link>
			<description>Description 2</description>
			<pubDate>Tue, 02 Jan 2024 10:00:00 GMT</pubDate>
		</item>
		<item>
			<title>Article from Jan 3</title>
			<link>https://example.com/article3</link>
			<description>Description 3</description>
			<pubDate>Wed, 03 Jan 2024 10:00:00 GMT</pubDate>
		</item>
		<item>
			<title>Article from Jan 4</title>
			<link>https://example.com/article4</link>
			<description>Description 4</description>
			<pubDate>Thu, 04 Jan 2024 10:00:00 GMT</pubDate>
		</item>
		<item>
			<title>Article from Jan 5</title>
			<link>https://example.com/article5</link>
			<description>Description 5</description>
			<pubDate>Fri, 05 Jan 2024 10:00:00 GMT</pubDate>
		</item>
	</channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validRSS))
	}))
	defer server.Close()

	t.Run("SyncModeNone - returns no articles", func(t *testing.T) {
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeNone, nil, nil)
		assert.NoError(t, err)
		assert.Empty(t, articles)
	})

	t.Run("SyncModeAll - returns all articles", func(t *testing.T) {
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeAll, nil, nil)
		assert.NoError(t, err)
		assert.Len(t, articles, 5)

		// Verify we got all the articles
		titles := make([]string, len(articles))
		for i, article := range articles {
			titles[i] = article.Title
		}
		assert.Contains(t, titles, "Article from Jan 1")
		assert.Contains(t, titles, "Article from Jan 5")
	})

	t.Run("SyncModeCount - returns specified number of most recent articles", func(t *testing.T) {
		count := 3
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeCount, &count, nil)
		assert.NoError(t, err)
		assert.Len(t, articles, 3)

		// Should be sorted by date, newest first
		assert.Equal(t, "Article from Jan 5", articles[0].Title)
		assert.Equal(t, "Article from Jan 4", articles[1].Title)
		assert.Equal(t, "Article from Jan 3", articles[2].Title)
	})

	t.Run("SyncModeCount - count larger than available articles", func(t *testing.T) {
		count := 10
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeCount, &count, nil)
		assert.NoError(t, err)
		assert.Len(t, articles, 5) // Should return all available articles
	})

	t.Run("SyncModeCount - invalid count (nil)", func(t *testing.T) {
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, "count", nil, nil)
		assert.NoError(t, err)
		assert.Empty(t, articles)
	})

	t.Run("SyncModeCount - invalid count (zero)", func(t *testing.T) {
		count := 0
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeCount, &count, nil)
		assert.NoError(t, err)
		assert.Empty(t, articles)
	})

	t.Run("SyncModeCount - invalid count (negative)", func(t *testing.T) {
		count := -5
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeCount, &count, nil)
		assert.NoError(t, err)
		assert.Empty(t, articles)
	})

	t.Run("SyncModeDateFrom - returns articles from specified date", func(t *testing.T) {
		// Filter from Jan 3, 2024 - should include Jan 3, 4, 5
		syncDate := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeDateFrom, nil, &syncDate)
		assert.NoError(t, err)
		assert.Len(t, articles, 3)

		// Verify we got the right articles
		titles := make([]string, len(articles))
		for i, article := range articles {
			titles[i] = article.Title
		}
		assert.Contains(t, titles, "Article from Jan 3")
		assert.Contains(t, titles, "Article from Jan 4")
		assert.Contains(t, titles, "Article from Jan 5")
		assert.NotContains(t, titles, "Article from Jan 1")
		assert.NotContains(t, titles, "Article from Jan 2")
	})

	t.Run("SyncModeDateFrom - exact date match included", func(t *testing.T) {
		// Filter from exactly Jan 3, 2024 10:00:00 GMT - should include the exact match
		syncDate := time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC)
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeDateFrom, nil, &syncDate)
		assert.NoError(t, err)
		assert.Len(t, articles, 3)

		// Find the Jan 3 article
		var jan3Article *rss.Article
		for _, article := range articles {
			if article.Title == "Article from Jan 3" {
				jan3Article = &article
				break
			}
		}
		assert.NotNil(t, jan3Article, "Article from Jan 3 should be included")
	})

	t.Run("SyncModeDateFrom - future date returns no articles", func(t *testing.T) {
		// Filter from future date - should return no articles
		syncDate := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeDateFrom, nil, &syncDate)
		assert.NoError(t, err)
		assert.Empty(t, articles)
	})

	t.Run("SyncModeDateFrom - invalid date (nil)", func(t *testing.T) {
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, "date_from", nil, nil)
		assert.NoError(t, err)
		assert.Empty(t, articles)
	})

	t.Run("Unknown sync mode - returns no articles", func(t *testing.T) {
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, "unknown_mode", nil, nil)
		assert.NoError(t, err)
		assert.Empty(t, articles)
	})

	t.Run("Error in base FetchAndParse propagates", func(t *testing.T) {
		articles, err := processor.FetchAndParseWithSyncOptions("invalid-url", models.SyncModeAll, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, articles)
		assert.Contains(t, err.Error(), "FetchAndParse failed")
	})
}

func TestProcessor_FetchAndParseWithSyncOptions_ArticlesWithoutDates(t *testing.T) {
	processor := rss.NewProcessor()

	// Create RSS with some articles having nil dates
	mixedDateRSS := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Mixed Date Feed</title>
		<description>RSS Feed with mixed date scenarios</description>
		<link>https://example.com</link>
		<item>
			<title>Article with date</title>
			<link>https://example.com/with-date</link>
			<description>Has publication date</description>
			<pubDate>Wed, 03 Jan 2024 10:00:00 GMT</pubDate>
		</item>
		<item>
			<title>Article without date</title>
			<link>https://example.com/no-date</link>
			<description>No publication date</description>
		</item>
	</channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mixedDateRSS))
	}))
	defer server.Close()

	t.Run("SyncModeDateFrom - articles without dates get current time", func(t *testing.T) {
		// Use a future date so articles with actual past dates are excluded
		syncDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeDateFrom, nil, &syncDate)
		assert.NoError(t, err)
		// The article without date gets current time (which is after 2024-06-01), so it should be included
		// The article with date (2024-01-03) should be excluded
		assert.Len(t, articles, 1)
		assert.Equal(t, "Article without date", articles[0].Title)
	})

	t.Run("SyncModeCount - handles mixed date scenarios", func(t *testing.T) {
		count := 2
		articles, err := processor.FetchAndParseWithSyncOptions(server.URL, models.SyncModeCount, &count, nil)
		assert.NoError(t, err)
		assert.Len(t, articles, 2)

		// Article with date should come first (newest first sorting)
		// Since articles without dates get current time, they might be sorted differently
		// but we should still get both articles
		titles := make([]string, len(articles))
		for i, article := range articles {
			titles[i] = article.Title
		}
		assert.Contains(t, titles, "Article with date")
		assert.Contains(t, titles, "Article without date")
	})
}

func BenchmarkProcessor_FetchAndParse(b *testing.B) {
	// Create a test server with RSS content
	validRSS := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Benchmark Feed</title>
		<description>RSS Feed for benchmarking</description>
		<link>https://example.com</link>`

	// Add many items for a more realistic benchmark
	for i := 0; i < 100; i++ {
		validRSS += fmt.Sprintf(`
		<item>
			<title>Article %d</title>
			<link>https://example.com/article%d</link>
			<description>Description of article %d</description>
			<pubDate>Mon, 01 Jan 2024 10:00:00 GMT</pubDate>
		</item>`, i, i, i)
	}

	validRSS += `
	</channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(validRSS))
	}))
	defer server.Close()

	processor := rss.NewProcessor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		articles, err := processor.FetchAndParse(server.URL)
		if err != nil {
			b.Fatal(err)
		}
		if len(articles) != 100 {
			b.Fatalf("Expected 100 articles, got %d", len(articles))
		}
	}
}
