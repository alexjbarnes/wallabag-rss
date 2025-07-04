package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"wallabag-rss-tool/pkg/config"
	"wallabag-rss-tool/pkg/database"
	"wallabag-rss-tool/pkg/logging"
	"wallabag-rss-tool/pkg/models"
	"wallabag-rss-tool/pkg/wallabag"
	"wallabag-rss-tool/pkg/worker"
	"wallabag-rss-tool/views"
)

const localhostIP = "localhost"

// Server holds the HTTP server and its dependencies.
type Server struct {
	store          database.Storer
	wallabagClient wallabag.Clienter
	worker         *worker.Worker
	csrfManager    *CSRFManager
}

// NewServer creates a new Server instance.
func NewServer(store database.Storer, wallabagClient wallabag.Clienter, worker *worker.Worker) *Server {
	return &Server{
		store:          store,
		wallabagClient: wallabagClient,
		worker:         worker,
		csrfManager:    NewCSRFManager(),
	}
}

// GetLocalIP returns the local IP address without external connections
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return localhostIP
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return localhostIP
}

// Start configures and starts the HTTP server.
func (s *Server) Start(port string) error {
	// Create secure HTTP server with timeouts
	mux := http.NewServeMux()
	
	
	mux.HandleFunc("/", s.AddSecurityHeaders(s.HandleIndex))
	mux.HandleFunc("/feeds/", s.AddSecurityHeaders(s.csrfProtection(s.handleFeeds)))
	mux.HandleFunc("/feeds/edit/", s.AddSecurityHeaders(s.handleEditFeed))
	mux.HandleFunc("/feeds/row/", s.AddSecurityHeaders(s.handleFeedRow))
	mux.HandleFunc("/articles", s.AddSecurityHeaders(s.handleArticles))
	mux.HandleFunc("/settings", s.AddSecurityHeaders(s.handleSettings))
	mux.HandleFunc("/sync", s.AddSecurityHeaders(s.csrfProtection(s.handleSync)))
	mux.HandleFunc("/settings/poll-interval", s.AddSecurityHeaders(s.csrfProtection(s.handleUpdateDefaultPollInterval)))

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        mux,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	ip := GetLocalIP()
	logging.Info("Server starting", "ip", ip, "port", port, "url", fmt.Sprintf("http://%s:%s", ip, port))

	return server.ListenAndServe()
}

// AddSecurityHeaders adds security headers to HTTP responses
func (s *Server) AddSecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Security headers
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("X-Frame-Options", "DENY")
		writer.Header().Set("X-XSS-Protection", "1; mode=block")
		writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		writer.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; script-src 'self' 'unsafe-inline' https://unpkg.com https://cdn.jsdelivr.net")

		// Call the next handler
		next.ServeHTTP(writer, request)
	})
}

func (s *Server) HandleIndex(writer http.ResponseWriter, request *http.Request) {
	data := views.PageData{Title: "Wallabag RSS Tool", CSRFToken: s.getCSRFToken()}
	if err := views.Index(data).Render(request.Context(), writer); err != nil {
		http.Error(writer, "Failed to render template", http.StatusInternalServerError)
	}
}

func (s *Server) handleFeeds(writer http.ResponseWriter, request *http.Request) {
	// Check if this is a request for a specific feed (has ID in path)
	// Path will be either "/feeds/" (collection) or "/feeds/123" (specific feed)
	if request.URL.Path != "/feeds/" && len(request.URL.Path) > len("/feeds/") {
		// This is a request for a specific feed: /feeds/{id}
		switch request.Method {
		case "PUT":
			s.handleFeedsPut(writer, request)
		case "DELETE":
			s.handleFeedsDelete(writer, request)
		default:
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// This is a request for the feeds collection: /feeds/ or /feeds
	switch request.Method {
	case http.MethodGet:
		s.handleFeedsGet(writer, request)
	case http.MethodPost:
		s.handleFeedsPost(writer, request)
	default:
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleFeedsGet handles GET requests for feeds listing
func (s *Server) handleFeedsGet(writer http.ResponseWriter, request *http.Request) {
	feeds, err := s.store.GetFeeds(request.Context())
	if err != nil {
		logging.Error("Failed to get feeds", "error", fmt.Errorf("store.GetFeedsWithContext: %w", err))
		http.Error(writer, "Failed to get feeds", http.StatusInternalServerError)

		return
	}

	defaultPollInterval := s.getDefaultPollIntervalWithFallback(request.Context())
	data := views.FeedsData{
		PageData:            views.PageData{Title: "Manage RSS Feeds", CSRFToken: s.getCSRFToken()},
		Feeds:               feeds,
		DefaultPollInterval: defaultPollInterval,
	}

	if err := views.Feeds(data).Render(request.Context(), writer); err != nil {
		http.Error(writer, "Failed to render feeds template", http.StatusInternalServerError)
	}
}

// handleFeedsPost handles POST requests for creating new feeds
func (s *Server) handleFeedsPost(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, "Failed to parse form", http.StatusBadRequest)

		return
	}

	feed := s.parseFeedFromForm(request)
	id, err := s.store.InsertFeed(request.Context(), &feed)
	if err != nil {
		logging.Error("Failed to insert feed",
			"error", fmt.Errorf("store.InsertFeed: %w", err),
			"feed_name", feed.Name,
			"feed_url", feed.URL)
		http.Error(writer, "Failed to add feed", http.StatusInternalServerError)

		return
	}

	feed.ID = int(id)
	logging.Info("Feed added successfully",
		"feed_id", feed.ID,
		"feed_name", feed.Name,
		"feed_url", feed.URL,
		"sync_mode", feed.SyncMode)

	// Queue the new feed for immediate processing
	s.worker.QueueFeedForImmediate(feed.ID)

	s.renderFeedRow(writer, request, &feed)
}

// handleFeedsPut handles PUT requests for updating feeds
func (s *Server) handleFeedsPut(writer http.ResponseWriter, request *http.Request) {
	id, err := s.ExtractFeedIDFromPath(request.URL.Path)
	if err != nil {
		http.Error(writer, "Invalid feed ID", http.StatusBadRequest)
		return
	}

	existingFeed, err := s.store.GetFeedByID(request.Context(), id)
	if err != nil {
		logging.Error("Failed to get existing feed for update",
			"error", fmt.Errorf("store.GetFeedByID: %w", err),
			"feed_id", id)
		http.Error(writer, "Feed not found", http.StatusNotFound)
		return
	}

	if err := request.ParseForm(); err != nil {
		http.Error(writer, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Parse only the editable fields from the edit form
	formValues := s.ExtractFormValues(request)
	s.LogFormValues(&formValues)

	pollInterval, pollIntervalUnit := s.ParsePollInterval(formValues.PollIntervalStr, formValues.PollIntervalUnitStr)

	// Create updated feed preserving sync settings from existing feed
	feed := *existingFeed
	feed.Name = formValues.Name
	feed.URL = formValues.URL
	feed.SetPollInterval(pollInterval, pollIntervalUnit)

	if err := s.store.UpdateFeed(request.Context(), &feed); err != nil {
		logging.Error("Failed to update feed",
			"error", fmt.Errorf("store.UpdateFeed: %w", err),
			"feed_id", feed.ID,
			"feed_name", feed.Name)
		http.Error(writer, "Failed to update feed", http.StatusInternalServerError)
		return
	}

	logging.Info("Feed updated successfully",
		"feed_id", feed.ID,
		"feed_name", feed.Name,
		"feed_url", feed.URL)

	// Queue the updated feed for immediate re-sync if URL changed
	if existingFeed.URL != feed.URL {
		s.worker.QueueFeedForImmediate(feed.ID)
		logging.Info("Feed queued for re-sync due to URL change", "feed_id", feed.ID)
	}

	s.renderFeedRow(writer, request, &feed)
}

// handleFeedsDelete handles DELETE requests for removing feeds
func (s *Server) handleFeedsDelete(writer http.ResponseWriter, request *http.Request) {
	id, err := s.ExtractFeedIDFromPath(request.URL.Path)
	if err != nil {
		http.Error(writer, "Invalid feed ID", http.StatusBadRequest)

		return
	}

	if err := s.store.DeleteFeed(request.Context(), id); err != nil {
		logging.Error("Failed to delete feed",
			"error", fmt.Errorf("store.DeleteFeed: %w", err),
			"feed_id", id)
		http.Error(writer, "Failed to delete feed", http.StatusInternalServerError)

		return
	}

	logging.Info("Feed deleted successfully", "feed_id", id)
	writer.WriteHeader(http.StatusOK)
}

// Helper methods

// getDefaultPollIntervalWithFallback gets the default poll interval or returns fallback
func (s *Server) getDefaultPollIntervalWithFallback(ctx context.Context) int {
	defaultPollInterval, err := s.store.GetDefaultPollInterval(ctx)
	if err != nil {
		logging.Warn("Error getting default poll interval, using fallback",
			"error", fmt.Errorf("store.GetDefaultPollInterval: %w", err),
			"fallback_minutes", 60)

		return 60
	}

	return defaultPollInterval
}

// ExtractFeedIDFromPath extracts feed ID from URL path
func (s *Server) ExtractFeedIDFromPath(path string) (int, error) {
	idStr := path[len("/feeds/"):]

	return strconv.Atoi(idStr)
}

// parseFeedFromForm parses form data into a Feed struct
func (s *Server) parseFeedFromForm(request *http.Request) models.Feed {
	formValues := s.ExtractFormValues(request)
	s.LogFormValues(&formValues)

	pollInterval, pollIntervalUnit := s.ParsePollInterval(formValues.PollIntervalStr, formValues.PollIntervalUnitStr)
	syncMode := s.ParseSyncMode(formValues.SyncModeStr)
	syncCount := s.ParseSyncCount(formValues.SyncCountStr, syncMode)
	syncDateFrom := s.ParseSyncDateFrom(formValues.SyncDateFromStr, syncMode)

	feed := models.Feed{
		Name:            formValues.Name,
		URL:             formValues.URL,
		SyncMode:        syncMode,
		SyncCount:       syncCount,
		SyncDateFrom:    syncDateFrom,
		InitialSyncDone: false,
	}

	feed.SetPollInterval(pollInterval, pollIntervalUnit)

	logging.Info("DEBUG: Feed created",
		"poll_interval", feed.PollInterval,
		"poll_interval_unit", feed.PollIntervalUnit,
		"sync_mode", feed.SyncMode,
		"sync_count", syncCount,
		"sync_date_from", syncDateFrom)

	return feed
}

type FormValues struct {
	Name                string
	URL                 string
	PollIntervalStr     string
	PollIntervalUnitStr string
	SyncModeStr         string
	SyncCountStr        string
	SyncDateFromStr     string
}

func (s *Server) ExtractFormValues(request *http.Request) FormValues {
	return FormValues{
		Name:                request.FormValue("name"),
		URL:                 request.FormValue("url"),
		PollIntervalStr:     request.FormValue("poll_interval"),
		PollIntervalUnitStr: request.FormValue("poll_interval_unit"),
		SyncModeStr:         request.FormValue("sync_mode"),
		SyncCountStr:        request.FormValue("sync_count"),
		SyncDateFromStr:     request.FormValue("sync_date_from"),
	}
}

func (s *Server) LogFormValues(fv *FormValues) {
	logging.Info("DEBUG: Form values received",
		"name", fv.Name,
		"url", fv.URL,
		"poll_interval", fv.PollIntervalStr,
		"poll_interval_unit", fv.PollIntervalUnitStr,
		"sync_mode", fv.SyncModeStr,
		"sync_count", fv.SyncCountStr,
		"sync_date_from", fv.SyncDateFromStr)
}

func (s *Server) ParsePollInterval(pollIntervalStr, pollIntervalUnitStr string) (int, models.TimeUnit) {
	pollInterval, err := strconv.Atoi(pollIntervalStr)
	if err != nil {
		logging.Info("DEBUG: Poll interval conversion failed", "value", pollIntervalStr, "error", err)
		pollInterval = 0
	}

	pollIntervalUnit := models.TimeUnit(pollIntervalUnitStr)
	if pollIntervalUnit == "" {
		pollIntervalUnit = models.TimeUnitDays
	}

	return pollInterval, pollIntervalUnit
}

func (s *Server) ParseSyncMode(syncModeStr string) models.SyncMode {
	if syncModeStr == "" {
		syncModeStr = "none"
	}

	return models.SyncMode(syncModeStr)
}

func (s *Server) ParseSyncCount(syncCountStr string, syncMode models.SyncMode) *int {
	if syncCountStr != "" && syncMode == models.SyncModeCount {
		if count, err := strconv.Atoi(syncCountStr); err == nil && count > 0 {
			logging.Info("DEBUG: Sync count parsed", "value", count)
			return &count
		}
		logging.Info("DEBUG: Sync count conversion failed", "value", syncCountStr)
	}

	return nil
}

func (s *Server) ParseSyncDateFrom(syncDateFromStr string, syncMode models.SyncMode) *time.Time {
	if syncDateFromStr != "" && syncMode == models.SyncModeDateFrom {
		if date, err := time.Parse("2006-01-02", syncDateFromStr); err == nil {
			logging.Info("DEBUG: Sync date parsed", "value", date)
			return &date
		}
		logging.Info("DEBUG: Sync date conversion failed", "value", syncDateFromStr)
	}

	return nil
}

// renderFeedRow renders a feed row for HTMX responses
func (s *Server) renderFeedRow(writer http.ResponseWriter, request *http.Request, feed *models.Feed) {
	defaultPollInterval := s.getDefaultPollIntervalWithFallback(request.Context())
	if err := views.FeedRow(*feed, defaultPollInterval, s.getCSRFToken()).Render(request.Context(), writer); err != nil {
		http.Error(writer, "Failed to render feed row", http.StatusInternalServerError)
	}
}

func (s *Server) handleEditFeed(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}
	idStr := request.URL.Path[len("/feeds/edit/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid feed ID", http.StatusBadRequest)

		return
	}
	feed, err := s.store.GetFeedByID(request.Context(), id)
	if err != nil {
		http.Error(writer, "Feed not found", http.StatusNotFound)

		return
	}

	defaultPollInterval, err := s.store.GetDefaultPollInterval(request.Context())
	if err != nil {
		logging.Warn("Error getting default poll interval for edit form, using fallback",
			"error", fmt.Errorf("store.GetDefaultPollInterval: %w", err),
			"fallback_minutes", 60)
		defaultPollInterval = 60 // fallback to 60 minutes
	}

	data := views.FeedEditData{
		Feed:                *feed,
		DefaultPollInterval: defaultPollInterval,
		CSRFToken:           s.getCSRFToken(),
	}
	if err := views.FeedEditForm(data).Render(request.Context(), writer); err != nil {
		http.Error(writer, "Failed to render edit form", http.StatusInternalServerError)
	}
}

func (s *Server) handleFeedRow(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}
	idStr := request.URL.Path[len("/feeds/row/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid feed ID", http.StatusBadRequest)

		return
	}
	feed, err := s.store.GetFeedByID(request.Context(), id)
	if err != nil {
		http.Error(writer, "Feed not found", http.StatusNotFound)

		return
	}

	defaultPollInterval, err := s.store.GetDefaultPollInterval(request.Context())
	if err != nil {
		logging.Warn("Error getting default poll interval for feed row, using fallback",
			"error", fmt.Errorf("store.GetDefaultPollInterval: %w", err),
			"fallback_minutes", 60)
		defaultPollInterval = 60 // fallback to 60 minutes
	}

	if err := views.FeedRow(*feed, defaultPollInterval, s.getCSRFToken()).Render(request.Context(), writer); err != nil {
		http.Error(writer, "Failed to render feed row", http.StatusInternalServerError)
	}
}

func (s *Server) handleArticles(writer http.ResponseWriter, request *http.Request) {
	articles, err := s.store.GetArticles(request.Context())
	if err != nil {
		http.Error(writer, "Failed to get articles", http.StatusInternalServerError)

		return
	}
	data := views.ArticlesData{
		PageData: views.PageData{Title: "Processed Articles", CSRFToken: s.getCSRFToken()},
		Articles: articles,
	}
	if err := views.Articles(data).Render(request.Context(), writer); err != nil {
		http.Error(writer, "Failed to render articles", http.StatusInternalServerError)
	}
}

func (s *Server) handleSettings(writer http.ResponseWriter, request *http.Request) {
	wallabagConfigLoaded := true
	if _, err := config.LoadWallabagConfig(); err != nil {
		wallabagConfigLoaded = false
	}

	defaultPollInterval, err := s.store.GetDefaultPollInterval(request.Context())
	if err != nil {
		logging.Warn("Error getting default poll interval for settings page, using fallback",
			"error", fmt.Errorf("store.GetDefaultPollInterval: %w", err),
			"fallback_minutes", 60)
		defaultPollInterval = 60 // Fallback
	}

	data := views.SettingsData{
		PageData:             views.PageData{Title: "Settings", CSRFToken: s.getCSRFToken()},
		WallabagConfigLoaded: wallabagConfigLoaded,
		DefaultPollInterval:  defaultPollInterval,
	}
	if err := views.Settings(data).Render(request.Context(), writer); err != nil {
		http.Error(writer, "Failed to render settings", http.StatusInternalServerError)
	}
}

func (s *Server) handleSync(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	logging.Info("Manual sync triggered by UI")

	// Queue all feeds for immediate processing
	if err := s.worker.QueueAllFeedsForImmediate(request.Context()); err != nil {
		logging.Error("Failed to queue feeds for sync", "error", err)
		http.Error(writer, "Failed to initiate sync", http.StatusInternalServerError)

		return
	}

	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte("Sync initiated.")); err != nil {
		logging.Error("Failed to write sync response", "error", err)
	}
}

func (s *Server) handleUpdateDefaultPollInterval(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "PUT" {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := request.ParseForm(); err != nil {
		http.Error(writer, "Failed to parse form", http.StatusBadRequest)
		return
	}

	interval, unit, err := s.ParseDefaultPollIntervalForm(request)
	if err != nil {
		http.Error(writer, "Invalid poll interval", http.StatusBadRequest)
		return
	}

	intervalInMinutes := s.ConvertToMinutes(interval, unit)

	if err := s.store.UpdateDefaultPollInterval(request.Context(), intervalInMinutes); err != nil {
		logging.Error("Failed to update default poll interval",
			"error", fmt.Errorf("store.UpdateDefaultPollInterval: %w", err),
			"interval_minutes", intervalInMinutes)
		http.Error(writer, "Failed to update default poll interval", http.StatusInternalServerError)

		return
	}

	logging.Info("Default poll interval updated", "interval_value", interval, "unit", unit, "interval_minutes", intervalInMinutes)

	// Return properly formatted HTML for HTMX target replacement
	response := s.FormatPollIntervalResponse(intervalInMinutes)
	if _, err := fmt.Fprint(writer, response); err != nil {
		logging.Error("Failed to write poll interval response", "error", err)
	}
}

func (s *Server) ParseDefaultPollIntervalForm(request *http.Request) (int, models.TimeUnit, error) {
	intervalStr := request.FormValue("default_poll_interval")
	unitStr := request.FormValue("default_poll_interval_unit")
	
	interval, err := strconv.Atoi(intervalStr)
	if err != nil || interval < 1 {
		return 0, "", fmt.Errorf("invalid interval: %s", intervalStr)
	}

	unit := models.TimeUnit(unitStr)
	if unit == "" {
		unit = models.TimeUnitHours
	}

	return interval, unit, nil
}

func (s *Server) ConvertToMinutes(interval int, unit models.TimeUnit) int {
	switch unit {
	case models.TimeUnitMinutes:
		return interval
	case models.TimeUnitHours:
		return interval * 60
	case models.TimeUnitDays:
		return interval * 60 * 24
	default:
		return interval * 60 // default to hours
	}
}

func (s *Server) FormatPollIntervalResponse(intervalInMinutes int) string {
	var display string
	
	switch {
	case intervalInMinutes == 1440:
		display = "1 day"
	case intervalInMinutes == 60:
		display = "1 hour"
	case intervalInMinutes%1440 == 0:
		display = fmt.Sprintf("%d days", intervalInMinutes/1440)
	case intervalInMinutes%60 == 0:
		display = fmt.Sprintf("%d hours", intervalInMinutes/60)
	default:
		display = fmt.Sprintf("%d minutes", intervalInMinutes)
	}
	
	return fmt.Sprintf(`<span id="default-poll-interval-display">%s</span>`, display)
}

// EqualIntPointers compares two int pointers for equality
func EqualIntPointers(first, second *int) bool {
	if first == nil && second == nil {
		return true
	}
	if first == nil || second == nil {
		return false
	}

	return *first == *second
}

// EqualTimePointers compares two time pointers for equality
func EqualTimePointers(first, second *time.Time) bool {
	if first == nil && second == nil {
		return true
	}
	if first == nil || second == nil {
		return false
	}

	return first.Equal(*second)
}

