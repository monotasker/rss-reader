package shared

import (
	"github.com/monotasker/rss-reader/internal/domain"
	"github.com/monotasker/rss-reader/internal/store"
)

// AppState holds the current state of the TUI
type AppState struct {
	Mode          string // the current UI mode: item_list, item_detail, feeds_list, feed_add, tags_list, tag_add, tag_assign, tag_parent
	Feeds         []domain.Feed
	SelectedFeeds []string
	Items         []domain.Item // the current filtered page of items (for display)
	SelectedItems []string      // the UUIDs of any selected items (for single selection, this is length 1)
	Tags          []domain.Tag
	SelectedTags  []string
	Filter        store.ItemFilter
	Page          int
	PageSize      int
}
