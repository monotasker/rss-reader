// Package service contains core application services to support
// both the CLI and the REST API interfaces
package service

import (
	"context"
	"errors"
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
	store   *store.Store
	fetcher *feed.Fetcher
	logger  *slog.Logger
}

// New initializes an App with its dependencies.
func New(store *store.Store, logger *slog.Logger) (*App, error) {
	fetcher, err := feed.NewFetcher()
	if err != nil {
		return nil, fmt.Errorf("create fetcher in App: %w", err)
	}
	return &App{store: store, fetcher: fetcher, logger: logger}, nil
}

// AddFeed subscribes to a feed URL: it fetches once to validate
// the feed and learn its title, then stores it.
func (app *App) AddFeed(ctx context.Context, url string) (domain.Feed, error) {
	if err := validateURL(url); err != nil {
		return domain.Feed{}, err
	}
	body, err := app.fetcher.Fetch(ctx, url)
	if err != nil {
		return domain.Feed{}, fmt.Errorf("add feed %s: %w", url, err)
	}
	parsed, err := feed.ParseRSS(ctx, body)
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
	if err := app.store.InsertFeed(ctx, feed); err != nil {
		return domain.Feed{}, err
	}
	app.logger.Info("feed added", "url", url, "title", feed.Title)
	return feed, nil
}

func (app *App) RemoveFeed(ctx context.Context, ref string) (domain.Feed, error) {
	feed, err := app.resolveFeed(ctx, ref)
	if err != nil {
		return domain.Feed{}, fmt.Errorf("retrieve feed %s for removal: %w", ref, err)
	}
	if err := app.store.DeleteFeed(ctx, feed.URL); err != nil {
		return domain.Feed{}, err
	}
	app.logger.Info("feed removed", "id", feed.ID, "url", feed.URL)
	return feed, nil
}

// PreviewFeed fetches and parses the URL for a feed.
// The function returns the parsed feed content without
// saving anything.
func (app *App) PreviewFeed(ctx context.Context, url string) (feed.ParsedFeed, error) {
	body, err := app.fetcher.Fetch(ctx, url)
	if err != nil {
		return feed.ParsedFeed{}, err
	}
	return feed.ParseRSS(ctx, body)
}

// ListFeeds returns all subscribed feeds.
func (app *App) ListFeeds(ctx context.Context) ([]domain.Feed, error) {
	return app.store.ListFeeds(ctx)
}

// resolveFeed finds a feed by URL first, then by ID.
func (app *App) resolveFeed(ctx context.Context, ref string) (domain.Feed, error) {
	f, err := app.store.GetFeedByURL(ctx, ref)
	if err == nil {
		return f, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return domain.Feed{}, err
	}
	feed, err := app.store.GetFeed(ctx, ref)
	if err != nil {
		return domain.Feed{}, err
	}
	return feed, nil
}
