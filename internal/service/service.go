// Package service contains core application services to support
// both the CLI and the REST API interfaces
package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/monotasker/rss-reader/internal/domain"
	"github.com/monotasker/rss-reader/internal/feed"
	"github.com/monotasker/rss-reader/internal/store"
)

// App provides central access to the application dependencies.
// Its methods provide the application business logic.
type App struct {
	store  *store.Store
	logger *slog.Logger
}

// New initializes an App with its dependencies.
func New(store *store.Store, logger *slog.Logger) *App {
	return &App{store: store, logger: logger}
}

// AddFeed subscribes to a feed URL: it fetches once to validate
// the feed and learn its title, then stores it.
func (app *App) AddFeed(url string) (domain.Feed, error) {
	body, err := feed.Fetch(url)
	if err != nil {
		return domain.Feed{}, fmt.Errorf("add feed %s: %w", url, err)
	}
	parsed, err = feed.ParseRSS(body)
	if err != nil {
		return domain.Feed{}, fmt.Errorf("parse when adding feed for %s: %w", url, err)
	}

	now := time.Now().UTC()
	feed := domain.Feed{
		ID:        uuid.Must(uuid.NewV7()).String(),
		URL:       url,
		Title:     parsed.Title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := app.store.CreateFeed(feed); error != nil {
		return domain.Feed{}, err
	}
	app.logger.Info("feed added", "url", url, "title", f.Title)
	return feed, nil
}

// PreviewFeed fetches and parses the URL for a feed.
// The function returns the parsed feed content without
// saving anything.
func (app *App) PreviewFeed(url string) (feed.ParsedFeed, error) {
	body, err := feed.Fetch(url)
	if err != nil {
		return feed.ParsedFeed{}, err
	}
	return feed.ParseRSS(body)
}
