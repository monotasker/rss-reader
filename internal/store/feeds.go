package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monotasker/rss-reader/internal/domain"
)

const timeFormat = time.RFC3339

func fmtTime(t time.Time) string { return t.UTC().Format(timeFormat) }

// fmtTimePtr convertys an optional time to an optional string
func fmtTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	timestring := fmtTime(*t)
	return &timestring
}

// parseTimePtr converts an optional string back to time.Time
func parseTimePtr(str *string) (*time.Time, error) {
	if str == nil {
		return nil, nil
	}
	timeObject, err := time.Parse(timeFormat, *str)
	if err != nil {
		return nil, err
	}
	return &timeObject, nil
}

// isUniqueViolation detects SQLite's unique constraint failure. The Go driver
// doesn't export a typed error for  this, so we match SQLite's returned error
// message string.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// CreateFeed inserts feeds table row. Returns ErrDuplicate if the URL is already
// in an existing row.
func (store *Store) InsertFeed(feed domain.Feed) error {
	_, err := store.db.Exec(`
		INSERT INTO feeds (id, url, title, last_fetched, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		feed.ID, feed.URL, feed.Title, fmtTimePtr(feed.LastFetched),
		fmtTime(feed.CreatedAt), fmtTime(feed.UpdatedAt),
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

func scanFeed(row rowScanner) (domain.Feed, error) {
	var feed domain.Feed
	var lastFetched *string
	var created, updated string
	err := row.Scan(&feed.ID, &feed.URL, &feed.Title, &lastFetched, &created, &updated)

	if err != nil {
		return domain.Feed{}, err
	}
	if feed.LastFetched, err = parseTimePtr(lastFetched); err != nil {
		return domain.Feed{}, fmt.Errorf("bad last_fetched: %w", err)
	}
	if feed.CreatedAt, err = time.Parse(timeFormat, created); err != nil {
		return domain.Feed{}, fmt.Errorf("bad created_at: %w", err)
	}
	if feed.UpdatedAt, err = time.Parse(timeFormat, updated); err != nil {
		return domain.Feed{}, fmt.Errorf("bad updated_at: %w", err)
	}

	return feed, nil
}

const feedCols = `id, url, title, last_fetched, created_at, updated_at`

// GetFeed returns a domain.Feed for the given id, or ErrNotFound.
func (store *Store) GetFeed(id string) (domain.Feed, error) {
	row := store.db.QueryRow(`SELECT `+feedCols+` FROM feeds WHERE id = ?`, id)
	feed, err := scanFeed(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Feed{}, fmt.Errorf("feed %s: %w", id, ErrNotFound)
	}
	if err != nil {
		return domain.Feed{}, fmt.Errorf("feed %s: %w", id, err)
	}
	return feed, nil
}

// ListFeed returns domain.Feed instances for all feeds ordered by title.
func (store *Store) ListFeeds() ([]domain.Feed, error) {
	rows, err := store.db.Query(`SELECT` + feedCols + ` FROM feeds ORDER BY title COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("list feeds: %w", err)
	}
	defer rows.Close()

	var feeds []domain.Feed
	for rows.Next() {
		feed, err := scanFeed(rows)
		if err != nil {
			return nil, fmt.Errorf("list feeds scan: %w", err)
		}
		feeds = append(feeds, feed)
	}
	return feeds, rows.Err()
}

// GetFeedByURL returns a domain.Feed instance for the feed matching the url, or ErrNotFound.
func (store *Store) GetFeedByURL(url string) (domain.Feed, error) {
	row := store.db.QueryRow(`SELECT `+feedCols+` FROM feeds WHERE url = ?`, url)
	feed, err := scanFeed(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Feed{}, fmt.Errorf("feed %s: %w", url, ErrNotFound)
	}
	if err != nil {
		return domain.Feed{}, fmt.Errorf("feed %s: %w", url, err)
	}
	return feed, nil
}

func (store *Store) TouchFeedFetched(id, title string) (domain.Feed, error) {
	nowTime := time.Now().UTC()
	result, err := store.db.Exec(`
		UPDATE feeds SET title = ?, last_fetched = ?) 
		WHERE id = ?`,
		title, nowTime, id,
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

	feed, err := store.GetFeed(id)
	if err != nil {
		return domain.Feed{}, fmt.Errorf("get updated feed %s: %w", id, err)
	}

	return feed, nil
}

func (store *Store) DeleteFeed(url string) error {
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
