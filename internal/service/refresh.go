package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/monotasker/rss-reader/internal/domain"
	"github.com/monotasker/rss-reader/internal/feed"
)

// RefreshResult records the result of a refresh operation
type RefreshResult struct {
	Feed domain.Feed
	NewItems int
}

// RefreshFeed fetches one feed (by URL or ID) and stores its items.
func (app *App) RefreshFeed(ctx context.Context, ref string) (RefreshResult, error) {
	feed, err := app.resolveFeed(ctx, ref)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("cannot resolve feed: %w", err)
	}
	return app.refresh(ctx, feed)
}

// refresh updates an already-resolved feed
func (app *App) refresh(ctx context.Context, feed domain.Feed) (RefreshResult, error) {
	body, err := app.fetcher.Fetch(ctx, feed.URL)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("cannot fetch feed: %w", err)
	}
	parsed, err := feed.ParseRSS(ctx, body)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("cannot parse feed: %w", err)
	}

	now := time.Now()
	items := make([]domain.Item, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		items = append(items, domain.Item{
			ID: uuid.Must(uuid.NewV7()).String()
		})
	}
}
