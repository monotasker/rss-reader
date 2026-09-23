package store_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/monotasker/rss-reader/internal/domain"
	"github.com/monotasker/rss-reader/internal/store"
)

// newTestStore opens a fresh database in a per-test temp dir.
func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

// testFeed builds a valid feed with the given url.
func testFeed(url string) domain.Feed {
	now := time.Now()
	return domain.Feed{
		ID:        uuid.Must(uuid.NewV7()).String(),
		URL:       url,
		Title:     "Test Feed",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestCreateAndGetFeed(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	f := testFeed("https://example.com/feed")

	if err := st.InsertFeed(ctx, f); err != nil {
		t.Fatalf("insert feed: %v", err)
	}

	got, err := st.GetFeed(ctx, f.ID)
	if err != nil {
		t.Fatalf("get feed %v: %v", f.ID, err)
	}
	if got.URL != f.URL || got.Title != f.Title {
		t.Errorf("get feed returned %+v, want %+v", got, f)
	}
	// Creation time should remain what we originally sent in
	if !got.CreatedAt.Equal(f.CreatedAt.Truncate(time.Second)) {
		t.Errorf("retrieve CreatedAt time = %v, want %v", got.CreatedAt, f.CreatedAt)
	}
}

func TestCreateFeedDuplicate(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	f := testFeed("https://example.com/feed")

	if err := st.InsertFeed(ctx, f); err != nil {
		t.Fatalf("first inserted feed: %v", err)
	}
	err := st.InsertFeed(ctx, testFeed("https://example.com/feed")) // same URL
	if !errors.Is(err, store.ErrDuplicate) {
		t.Fatalf("second feed insert returned %v, want store.ErrDuplicate", err)
	}
}

func TestGetFeedNotFound(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	_, err := st.GetFeed(ctx, "no-such-id")
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("err = %v, want store.ErrNotFound", err)
	}
}

func TestListFeeds(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)

	feeds := []domain.Feed{
		testFeed("https://example.com/feed1"),
		testFeed("https://example.com/feed2"),
		testFeed("https://example.com/feed3"),
		testFeed("https://example.com/feed4"),
	}
	for _, f := range feeds {
		if err := st.InsertFeed(ctx, f); err != nil {
			t.Fatalf("insert feed %s: %v", f.URL, err)
		}
	}
	listed, err := st.ListFeeds(ctx)
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	if len(listed) != len(feeds) {
		t.Fatalf("ListFeeds returned %v, want %v", len(listed), len(feeds))
	}
	for i, f := range feeds {
		if listed[i].ID != f.ID {
			t.Errorf("listed feed ID = %v, want %v", listed[i].ID, f.ID)
		}
		if listed[i].URL != f.URL {
			t.Errorf("listed feed ID = %v, want %v", listed[i].URL, f.URL)
		}
		if listed[i].CreatedAt.Sub(f.CreatedAt.UTC()) > time.Second {
			t.Errorf("listed feed ID = %v, want %v", listed[i].CreatedAt, f.CreatedAt.UTC())
		}
	}
}
