package feed_test

import (
	"testing"
	"time"

	"github.com/monotasker/rss-reader/internal/feed"
)

const sampleRSSResponse = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
	  <title>Example Blog</title>
		<item>
			<title>First Post</title>
			<link>https://example.com/1</link>
			<guid>post-1</guid>
			<description>Hello.</description>
			<pubDate>Sat, 25 Jul 2026 10:00:00 GMT</pubDate>
		</item>
		<item>
			<title>No GUID Post</title>
			<link>https://example.com/2</link>
			<description>World.</description>
		</item>
	</channel>
</rss>`

func TestParseRSS(t *testing.T) {
	tests := []struct {
		name             string
		input            []byte
		title            string
		itemCount        int
		firstGUID        string
		firstPublishedAt time.Time
		secondGUID       string
	}{
		{
			name:             "secondNoGUID",
			input:            []byte(sampleRSSResponse),
			title:            "Example Blog",
			itemCount:        2,
			firstGUID:        "post-1",
			firstPublishedAt: time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC),
			secondGUID:       "https://example.com/2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := feed.ParseRSS(tt.input)
			if err != nil {
				t.Fatalf("ParseRSS: %v", err)
			}

			if got.Title != tt.title {
				t.Errorf("title = %q, want %q", got.Title, "Ex")
			}
			if len(got.Items) != tt.itemCount {
				t.Fatalf("len(items) = %d, want %d", len(got.Items), tt.itemCount)
			}

			first := got.Items[0]
			if first.GUID != tt.firstGUID {
				t.Errorf("first item GUID = %q, want %q", first.GUID, tt.firstGUID)
			}
			if !first.PublishedAt.Equal(tt.firstPublishedAt) {
				t.Errorf("first item PublishedAt = %v, want %v", *first.PublishedAt, tt.firstPublishedAt)
			}

			if second := got.Items[1]; second.GUID != tt.secondGUID {
				t.Errorf("second item GUID = %q, want %q", second.GUID, tt.secondGUID)
			}
		})
	}
}

func TestParseRSSRejectsGarbage(t *testing.T) {
	if _, err := feed.ParseRSS([]byte("this is not xml")); err == nil {
		t.Fatal("ParseRSS accepted garbage input")
	}
}
