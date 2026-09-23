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
	Feed     domain.Feed
	NewItems int
}

// RefreshFeed fetches one feed (by URL or ID) and stores its items.
func (app *App) RefreshFeed(ctx context.Context, ref string) (RefreshResult, error) {
	f, err := app.resolveFeed(ctx, ref)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("cannot resolve feed: %w", err)
	}
	return app.refresh(ctx, f)
}

// refresh updates an already-resolved feed
func (app *App) refresh(ctx context.Context, f domain.Feed) (RefreshResult, error) {
	body, err := app.fetcher.Fetch(ctx, f.URL)
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
			ID:          uuid.Must(uuid.NewV7()).String(),
			FeedID:      f.ID,
			GUID:        item.GUID,
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.PublishedAt,
			FetchedAt:   now,
		})
	}
	n, err := app.store.UpsertItems(ctx, f.ID, items)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("cannot update feed items for %s: %w", f.ID, err)
	}

	title := parsed.Title
	if title == "" {
		title = f.Title
	}
	touched, err := app.store.TouchFeedFetched(ctx, f.ID, title)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("cannot touch feed fetched for %s: %w", f.ID, err)
	}
	if touched.ID != f.ID {
		return RefreshResult{}, fmt.Errorf("feed ID mismatch after touching: %q vs %q", touched.ID, f.ID)
	}

	f.Title = title
	app.logger.Info("feed refreshed", "url", f.URL, "items", n, "total", len(items))
	return RefreshResult{Feed: f, NewItems: n}, nil
}
