package domain

import "time"

// Feed is a subscription that we periodically fetch
type Feed struct {
	ID          string // uuidv7
	URL         string
	Title       string
	LastFetched *time.Time // nil until we've fetched once
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Item is one item or entry from a feed
type Item struct {
	ID          string // uuidv7
	FeedID      string // uuidv7
	GUID        string
	Title       string
	Link        string
	PublishedAt *time.Time
	Summary     string
	Content     string
	Read        bool
	Bookmarked  bool
	Dismissed   bool
	FetchedAt   time.Time
	UpdatedAt   *time.Time
}
