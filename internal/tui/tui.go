/*
Package tui

# The TUI interface for rss-reader

Includes 4 views:
- Items view: The main UI entry point
  - Lists currently downloaded items
  - Filtered by selected feed(s), tag(s), and item state(s)
  - Item states may be read/unread, dismissed/undismissed, bookmarked/unbookmarked
  - Filters are added via the Feeds, Tags, or Filters overlay views
  - Item details can be opened as another overlay view
  - navigation:
  - vim bindings for movement,
  - space for select/deselect toggle,
  - r for refresh of feeds
  - enter or o to open item details
  - f to open feeds view
  - t to open tags view
  - x to toggle dismissed/undismissed
  - b to toggle bookmarked/unbookmarked
  - d to toggle read/unread
  - / to open filter view

- Feeds list view: Ui for listing/deleting/marking/filtering feeds
  - Lists currently subscribed feeds
  - Similar navigation and options to items view
  - a opens "feeds add" view as an additional overlay
  - x unsubscribes a feed instead of dismissing
  - esc returns to parent view (Items) and applies any feed selections to the items
    view filter

- Feeds add view: UI for adding feed
  - input box for entering URL (immediately focused)
  - enter submits the URL for subscribing
  - activity/progress indicator and result message (success or failure/error)
    displayed above input

- Tags list view: UI for listing/deleting/marking/filtering tags
  - Lists currently created tags, but as an expandable hierarchical tree
  - works just like the feeds view
  - a opens "tags create" view as an additional overlay
  - p with a tag selected opens "tags parent" view

- Tags create view: UI for creating a tag
  - overlay like feeds add view

- Tags parent view: UI for assigning
  - overlay with list of existing tags from which to select parents for the selected
    tag

- Item tag: UI to assign a tag to the currently selected feed(s) or item(s)
  - can apply to either feeds or items depending on the view from which it's opened

- Item details: UI for reading an item's title, summary, and/or full text (if available)
  - b opens the URL in the system browser
*/
package tui

import (
	"context"
	"fmt"
	"os"

	"github.com/monotasker/rss-reader/internal/service"
	"github.com/monotasker/rss-reader/internal/tui/shared"
	"github.com/monotasker/rss-reader/internal/tui/views"
	"golang.org/x/term"
)

// TUI is the main struct for the terminal UI
type TUI struct {
	App           *service.App
	State         *shared.AppState
	CurrentView   views.ViewState
	terminalState *term.State
}

func (t *TUI) Run() error {
	ctx := context.Background()
	t.init(ctx)
	defer t.cleanup()
	initialView, err := views.NewItemsView(t.App, t.State)
	t.CurrentView = initialView
	if err != nil {
		return fmt.Errorf("render initial NewItemsView: %w", err)
	}

	for {
		t.CurrentView.Render()

		key := ReadKey()
		newView, err := t.CurrentView.HandleKey(ctx, key)
		if err != nil {
			return err
		}
		switch newView {
		case nil:
			return nil // exit
		case t.CurrentView:
			// same view updated in place
		default:
			t.CurrentView = newView // change views
		}
	}
}

func (t *TUI) init(ctx context.Context) error {
	state, _ := term.MakeRaw(int(os.Stdin.Fd()))
	t.terminalState = state

	fmt.Print("\033[2J\033[H") // clear screen and move cursort to home
	return nil
}

func (t *TUI) cleanup() {
	defer term.Restore(int(os.Stdin.Fd()), t.terminalState)
}

// NewTUI creates a new TUI instance
func NewTUI(app *service.App) *TUI {
	return &TUI{
		App:   app,
		State: &shared.AppState{PageSize: 20},
	}
}
