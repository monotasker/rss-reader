package store

import (
	"context"
	"fmt"

	"github.com/monotasker/rss-reader/internal/domain"
	"github.com/monotasker/rss-reader/internal/utils"
)

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
