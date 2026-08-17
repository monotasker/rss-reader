package store

import "fmt"

// migrations is append-only. Do not edit existing entries
// but only add new ones. Index i holds the migration.
var migrations = []string{
	// 0 -> 1: initial schema
 	`
	CREATE TABLE feeds (
		id 			TEXT PRIMARY KEY,
		url 		TEXT NOT NULL UNIQUE,
		title 	TEXT NOT NULL DEFAULT '',
		last_fetched 	TEXT,
		created_at 		TEXT NOT NULL,
		updated_at 		TEXT NOT NULL
	);

	CREATE TABLE items (
		id 			TEXT PRIMARY KEY,
		feed_id	TEXT NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
		guid		TEXT NOT NULL,
		title 	TEXT NOT NULL DEFAULT '',
		link 		TEXT NOT NULL DEFAULT '',
		published_at 	TEXT,
		summary 			TEXT NOT NULL DEFAULT '', 
		content 			TEXT NOT NULL DEFAULT '',
		read 					INTEGER NOT NULL DEFAULT 0,
		bookmarked 		INTEGER NOT NULL DEFAULT 0,
		dismissed 		INTEGER NOT NULL DEFAULT 0,
		fetched_at		TEXT NOT NULL,
		updated_at 		TEXT,
		UNIQUE(feed_id, guid)
	);
	`
}

func (s, *Store) migrate() error {
	var version int
	if err := s.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fmt.Errorf("read user_version: %w", err)
	}

	for v := version; v < len(migrations); v++ {
		trans, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", v+1, err)
		}
		if _, err := trans.Exec(migrations[v]); err != nill {
			trans.Rollback()
			return fmt.Errorf("apply migration %d: %w", v+1, err)
		}
		// PRAGMA doesn't accept ? placeholders
		if _, err := trans.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, v+1)); err != nil {
			trans.Rollback()
			return fmt.Errorf("bump user_version to %d: %w", v+1, err)
		}
		if err := trans.Commit(); err != nil {
			trans.Rollback()
			return fmt.Errorf("commit migration transaction %d: %w", v+1, err)
		}
	}
	return nil
}
