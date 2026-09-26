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

const itemCols = `id, feed_id, guid, title, link, published_at, summary,
    content, read, bookmarked, dismissed, fetched_at, updated_at`

// UpsertItems inserts new items and refreshes existing ones.
// User state (read/bookmarked/dismissed) is never touched.
func (store *Store) UpsertItems(ctx context.Context, feedID string, items []domain.Item) (int, error) {
	session, err := store.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("upser items: begin: %w", err)
	}
	defer session.Rollback()

	existing := make(map[string]bool)
	rows, err := session.Query(`SELECT guid FROM items WHERE feed_id = ?`, feedID)
	if err != nil {
		return 0, fmt.Errorf("upsert items: query existing items: %w", err)
	}
	for rows.Next() {
		var guid string
		if err := rows.Scan(&guid); err != nil {
			rows.Close()
			return 0, fmt.Errorf("upsert items: scan existing items: %w", err)
		}
		existing[guid] = true
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("upsert items: guid rows: %w", err)
	}

	stmt, err := session.Prepare(`
		INSERT INTO items (id, feed_id, guid, title, link, published_at, summary, content, fetched_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (feed_id, guid) DO UPDATE SET
				title = excluded.title,
				link = excluded.link,
				summary = excluded.summary,
				updated_at = excluded.fetched_at`)
	if err != nil {
		return 0, fmt.Errorf("upsert items: prepare sql statement: %w", err)
	}
	defer stmt.Close()

	inserted := 0
	for _, item := range items {
		if _, err := stmt.Exec(item.ID, feedID, item.GUID, item.Title, item.Link,
			utils.FmtTimePointer(item.PublishedAt), item.Summary, item.Content,
			utils.FmtTime(item.FetchedAt)); err != nil {
			return 0, fmt.Errorf("upsert items: exec SQL statement for item %q: %w", item.GUID, err)
		}
		if !existing[item.GUID] {
			inserted++
			existing[item.GUID] = true
		}
	}

	if err := session.Commit(); err != nil {
		return 0, fmt.Errorf("upsert items: commit transaction: %w", err)
	}
	return inserted, nil
}

// scanItem reads one row into a domain.Item; works for *sql.Row and *sql.Rows
// via the rowScanner interface from feeds.go.
func scanItem(r rowScanner) (domain.Item, error) {
	var it domain.Item
	var published, updated *string
	var fetched string
	err := r.Scan(&it.ID, &it.FeedID, &it.GUID, &it.Title, &it.Link,
		&published, &it.Summary, &it.Content,
		&it.Read, &it.Bookmarked, &it.Dismissed,
		&fetched, &updated)
	if err != nil {
		return domain.Item{}, err
	}
	if it.PublishedAt, err = utils.ParseTimePointer(published); err != nil {
		return domain.Item{}, fmt.Errorf("bad published_at: %w", err)
	}
	if it.FetchedAt, err = utils.ParseTime(fetched); err != nil {
		return domain.Item{}, fmt.Errorf("bad fetched_at: %w", err)
	}
	if it.UpdatedAt, err = utils.ParseTimePointer(updated); err != nil {
		return domain.Item{}, fmt.Errorf("bad updated_at: %w", err)
	}
	return it, nil
}

// GetItem returns the item with the given full id, or ErrNotFound.
func (store *Store) GetItem(ctx context.Context, id string) (domain.Item, error) {
	row := store.db.QueryRow(`SELECT `+itemCols+` FROM items WHERE id = ?`, id)
	it, err := scanItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Item{}, fmt.Errorf("item %s: %w", id, ErrNotFound)
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("get item: %w", err)
	}
	return it, nil
}

// GetItemByPrefix returns the single item whose id starts with prefix.
// No match is ErrNotFound; several matches are ErrAmbiguous.
func (store *Store) GetItemByPrefix(ctx context.Context, prefix string) (domain.Item, error) {
	rows, err := store.db.Query(
		`SELECT `+itemCols+` FROM items WHERE id LIKE ? LIMIT 2`, prefix+"%")
	if err != nil {
		return domain.Item{}, fmt.Errorf("get item by prefix: %w", err)
	}
	defer rows.Close()

	var matches []domain.Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return domain.Item{}, fmt.Errorf("get item by prefix: %w", err)
		}
		matches = append(matches, it)
	}
	if err := rows.Err(); err != nil {
		return domain.Item{}, err
	}
	switch len(matches) {
	case 0:
		return domain.Item{}, fmt.Errorf("item %s: %w", prefix, ErrNotFound)
	case 1:
		return matches[0], nil
	default:
		return domain.Item{}, fmt.Errorf("item %s: %w", prefix, ErrAmbiguous)
	}
}

// ItemFilter selects which items ListItems returns. The zero value
// means: every feed, read or unread, bookmarked or not, dismissed excluded,
// newest 50.
type ItemFilter struct {
	FeedID           string // exact feed id; "" = all feeds
	UnreadOnly       bool
	BookmarkedOnly   bool
	IncludeDismissed bool
	Search           string // substring match on title or summary
	Limit, Offset    int    // Limit <= 0 means 50
}

// ListItems returns items matching f, newest first.
func (store *Store) ListItems(ctx context.Context, filter ItemFilter) ([]domain.Item, error) {
	var b strings.Builder
	b.WriteString(`SELECT ` + itemCols + ` FROM items WHERE 1=1`)
	var args []any

	if filter.FeedID != "" {
		b.WriteString(` AND feed_id = ?`)
		args = append(args, filter.FeedID)
	}
	if filter.UnreadOnly {
		b.WriteString(` AND read = 0`)
	}
	if filter.BookmarkedOnly {
		b.WriteString(` AND bookmarked = 1`)
	}
	if !filter.IncludeDismissed {
		b.WriteString(` AND dismissed = 0`)
	}
	if filter.Search != "" {
		b.WriteString(` AND (title LIKE ? OR summary LIKE ?)`)
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	b.WriteString(` ORDER BY published_at DESC NULLS LAST, fetched_at DESC
		LIMIT ? OFFSET ?`)
	args = append(args, limit, filter.Offset)

	rows, err := store.db.Query(b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var items []domain.Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("list items scan: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// SetItemRead sets the read flag, or returns ErrNotFound.
func (store *Store) SetItemRead(ctx context.Context, id string, v bool) error {
	res, err := store.db.Exec(
		`UPDATE items SET read = ?, updated_at = ? WHERE id = ?`,
		v, utils.FmtTime(time.Now()), id)
	if err != nil {
		return fmt.Errorf("set item read: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("set item read: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("item %s: %w", id, ErrNotFound)
	}
	return nil
}

// SetItemBookmarked sets the bookmarked flag, or returns ErrNotFound.
func (store *Store) SetItemBookmarked(ctx context.Context, id string, v bool) error {
	res, err := store.db.Exec(
		`UPDATE items SET bookmarked = ?, updated_at = ? WHERE id = ?`,
		v, utils.FmtTime(time.Now()), id)
	if err != nil {
		return fmt.Errorf("set item bookmarked: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("set item bookmarked: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("item %s: %w", id, ErrNotFound)
	}
	return nil
}

// SetItemDismissed sets the dismissed flag, or returns ErrNotFound.
func (store *Store) SetItemDismissed(ctx context.Context, id string, v bool) error {
	res, err := store.db.Exec(
		`UPDATE items SET dismissed = ?, updated_at = ? WHERE id = ?`,
		v, utils.FmtTime(time.Now()), id)
	if err != nil {
		return fmt.Errorf("set item dismissed: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("set item dismissed: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("item %s: %w", id, ErrNotFound)
	}
	return nil
}

// GetItemBySuffix returns the single item whose last 8 chars start with suffix.
// No match is ErrNotFound; several matches are ErrAmbiguous.
func (store *Store) GetItemBySuffix(ctx context.Context, suffix string) (domain.Item, error) {
	rows, err := store.db.Query(
		`SELECT `+itemCols+` FROM items WHERE SUBSTR(id, -8) LIKE ? LIMIT 2`, suffix+"%")
	if err != nil {
		return domain.Item{}, fmt.Errorf("get item by suffix: %w", err)
	}
	defer rows.Close()

	var matches []domain.Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return domain.Item{}, fmt.Errorf("get item by suffix: %w", err)
		}
		matches = append(matches, it)
	}
	if err := rows.Err(); err != nil {
		return domain.Item{}, err
	}
	switch len(matches) {
	case 0:
		return domain.Item{}, fmt.Errorf("item %s: %w", suffix, ErrNotFound)
	case 1:
		return matches[0], nil
	default:
		return domain.Item{}, fmt.Errorf("item %s: %w", suffix, ErrAmbiguous)
	}
}
