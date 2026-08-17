// Package service contains core application services to support
// both the CLI and the REST API interfaces
package service

import (
	"log/slog"

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
