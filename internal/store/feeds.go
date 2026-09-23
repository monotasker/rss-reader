package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monotasker/rss-reader/internal/domain"
	"github.com/monotasker/rss-reader/internal/utils"
)

// isUniqueViolation detects SQLite's unique constraint failure. The Go driver
// doesn't export a typed error for  this, so we match SQLite's returned error
// message string.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// InsertFeed inserts feeds table row. Returns ErrDuplicate if the URL is already
// in an existing row.
func (store *Store) InsertFeed(ctx context.Context, feed domain.Feed) error {
	_, err := store.db.Exec(`
		INSERT INTO feeds (id, url, title, last_fetched, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		feed.ID, feed.URL, feed.Title, utils.FmtTimePointer(feed.LastFetched),
		utils.FmtTime(feed.CreatedAt), utils.FmtTime(feed.UpdatedAt),
	)
	if isUniqueViolation(err) {
		return fmt.Errorf("feed %s: %w", feed.URL, ErrDuplicate)
	}
	if err != nil {
		return fmt.Errorf("insert feed: %w", err)
	}
	return nil
}

// scanFeed reads one row and creates a domain.Feed struct instance from the values. Works
// for both *sql.Row and *sql.Rows because both have a Scan method.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanFeed(ctx context.Context, row rowScanner) (domain.Feed, error) {
	var feed domain.Feed
	var lastFetched *string
	var created, updated string
	err := row.Scan(&feed.ID, &feed.URL, &feed.Title, &lastFetched, &created, &updated)

	if err != nil {
		return domain.Feed{}, err
	}
	if feed.LastFetched, err = utils.ParseTimePointer(lastFetched); err != nil {
		return domain.Feed{}, fmt.Errorf("bad last_fetched: %w", err)
	}
	if feed.CreatedAt, err = utils.ParseTime(created); err != nil {
		return domain.Feed{}, fmt.Errorf("bad created_at: %w", err)
	}
	if feed.UpdatedAt, err = utils.ParseTime(updated); err != nil {
		return domain.Feed{}, fmt.Errorf("bad updated_at: %w", err)
	}

	return feed, nil
}

const feedCols = `id, url, title, last_fetched, created_at, updated_at`

// GetFeed returns a domain.Feed for the given id, or ErrNotFound.
func (store *Store) GetFeed(ctx context.Context, id string) (domain.Feed, error) {
	row := store.db.QueryRow(`SELECT `+feedCols+` FROM feeds WHERE id = ?`, id)
	feed, err := scanFeed(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Feed{}, fmt.Errorf("feed %s: %w", id, ErrNotFound)
	}
	if err != nil {
		return domain.Feed{}, fmt.Errorf("feed %s: %w", id, err)
	}
	return feed, nil
}

// ListFeeds returns domain.Feed instances for all feeds ordered by title.
func (store *Store) ListFeeds(ctx context.Context) ([]domain.Feed, error) {
	rows, err := store.db.Query(`SELECT ` + feedCols + ` FROM feeds ORDER BY title COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("list feeds: %w", err)
	}
	defer rows.Close()

	var feeds []domain.Feed
	for rows.Next() {
		feed, err := scanFeed(ctx, rows)
		if err != nil {
			return nil, fmt.Errorf("list feeds scan: %w", err)
		}
		feeds = append(feeds, feed)
	}
	return feeds, rows.Err()
}

// GetFeedByURL returns a domain.Feed instance for the feed matching the url, or ErrNotFound.
func (store *Store) GetFeedByURL(ctx context.Context, url string) (domain.Feed, error) {
	row := store.db.QueryRow(`SELECT `+feedCols+` FROM feeds WHERE url = ?`, url)
	feed, err := scanFeed(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Feed{}, fmt.Errorf("feed %s: %w", url, ErrNotFound)
	}
	if err != nil {
		return domain.Feed{}, fmt.Errorf("feed %s: %w", url, err)
	}
	return feed, nil
}

func (store *Store) TouchFeedFetched(ctx context.Context, id string, title string) (domain.Feed, error) {
	nowTime := time.Now().UTC()
	result, err := store.db.Exec(`
		UPDATE feeds SET title = ?, last_fetched = ? 
		WHERE id = ?`,
		title, utils.FmtTime(nowTime), id,
	)
	if err != nil {
		return domain.Feed{}, fmt.Errorf("feed %s: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Feed{}, fmt.Errorf("check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.Feed{}, fmt.Errorf("feed %s: %w", id, ErrNotFound)
	}

	feed, err := store.GetFeed(ctx, id)
	if err != nil {
		return domain.Feed{}, fmt.Errorf("get updated feed %s: %w", id, err)
	}

	return feed, nil
}

func (store *Store) DeleteFeed(ctx context.Context, url string) error {
	result, err := store.db.Exec(`
		DELETE FROM feeds WHERE url = ?`,
		url,
	)
	if err != nil {
		return fmt.Errorf("delete feed %s: %w", url, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("feed %s: %w", url, ErrNotFound)
	}

	return nil
}
