package service

import (
	"context"
	"fmt"

	"github.com/monotasker/rss-reader/internal/domain"
	"github.com/monotasker/rss-reader/internal/store"
)

// ListItems returns stored items matching the filter. FeedID may be a
// feed URL or id; it is resolved before querying.
func (app *App) ListItems(ctx context.Context, filter store.ItemFilter) ([]domain.Item, error) {
	if filter.FeedID != "" {
		feed, err := app.resolveFeed(ctx, filter.FeedID)
		if err != nil {
			return nil, err
		}
		filter.FeedID = feed.ID
	}
	return app.store.ListItems(ctx, filter)
}

// GetItem finds one item by a unique id suffix.
func (app *App) GetItem(ctx context.Context, ref string) (domain.Item, error) {
	return app.store.GetItemBySuffix(ctx, ref)
}

// MarkItemRead flags the item as read and returns the updated item.
func (app *App) MarkItemRead(ctx context.Context, ref string) (domain.Item, error) {
	item, err := app.store.GetItemBySuffix(ctx, ref)
	if err != nil {
		return domain.Item{}, fmt.Errorf("mark item %s read: %w", ref, err)
	}
	if err := app.store.SetItemRead(ctx, item.ID, true); err != nil {
		return domain.Item{}, fmt.Errorf("mark item %s read: %w", ref, err)
	}
	return app.store.GetItem(ctx, item.ID)
}

// MarkItemUnread flags the item as unread and returns the updated item.
func (app *App) MarkItemUnread(ctx context.Context, ref string) (domain.Item, error) {
	item, err := app.store.GetItemBySuffix(ctx, ref)
	if err != nil {
		return domain.Item{}, fmt.Errorf("mark item %s unread: %w", ref, err)
	}
	if err := app.store.SetItemRead(ctx, item.ID, false); err != nil {
		return domain.Item{}, fmt.Errorf("mark item %s unread: %w", ref, err)
	}
	return app.store.GetItem(ctx, item.ID)
}

// BookmarkItem flags the item as bookmarked and returns the updated item.
func (app *App) BookmarkItem(ctx context.Context, ref string) (domain.Item, error) {
	item, err := app.store.GetItemBySuffix(ctx, ref)
	if err != nil {
		return domain.Item{}, fmt.Errorf("bookmark item %s: %w", ref, err)
	}
	if err := app.store.SetItemBookmarked(ctx, item.ID, true); err != nil {
		return domain.Item{}, fmt.Errorf("bookmark item %s: %w", ref, err)
	}
	return app.store.GetItem(ctx, item.ID)
}

// UnbookmarkItem flags the item as not bookmarked and returns the updated item.
func (app *App) UnbookmarkItem(ctx context.Context, ref string) (domain.Item, error) {
	item, err := app.store.GetItemBySuffix(ctx, ref)
	if err != nil {
		return domain.Item{}, fmt.Errorf("unbookmark item %s: %w", ref, err)
	}
	if err := app.store.SetItemBookmarked(ctx, item.ID, false); err != nil {
		return domain.Item{}, fmt.Errorf("unbookmark item %s: %w", ref, err)
	}
	return app.store.GetItem(ctx, item.ID)
}

// DismissItem flags the item as dismissed and returns the updated item.
func (app *App) DismissItem(ctx context.Context, ref string) (domain.Item, error) {
	item, err := app.store.GetItemBySuffix(ctx, ref)
	if err != nil {
		return domain.Item{}, fmt.Errorf("dismiss item %s: %w", ref, err)
	}
	if err := app.store.SetItemDismissed(ctx, item.ID, true); err != nil {
		return domain.Item{}, fmt.Errorf("dismiss item %s: %w", ref, err)
	}
	return app.store.GetItem(ctx, item.ID)
}

// UndismissItem flags the item as not dismissed and returns the updated item.
func (app *App) UndismissItem(ctx context.Context, ref string) (domain.Item, error) {
	item, err := app.store.GetItemBySuffix(ctx, ref)
	if err != nil {
		return domain.Item{}, fmt.Errorf("undismiss item %s: %w", ref, err)
	}
	if err := app.store.SetItemDismissed(ctx, item.ID, false); err != nil {
		return domain.Item{}, fmt.Errorf("undismiss item %s: %w", ref, err)
	}
	return app.store.GetItem(ctx, item.ID)
}
