package feed

import (
	"encoding/xml"
	"fmt"
	"time"
)

// rssDocument is the incoming wire format for RSS 2.0
type rssDocument struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// ParsedFeed is feed data decoded from XML, before we assign local ids
type ParsedFeed struct {
	Title string
	Items []ParsedItem
}

// ParsedItem is one entry from a feed document
type ParsedItem struct {
	GUID        string
	Title       string
	Link        string
	Summary     string
	PublishedAt *time.Time
}

// ParseRSS parses an RSS 2.0 feed into a ParsedFeed
func ParseRSS(data []byte) (ParsedFeed, error) {
	var doc rssDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		return ParsedFeed{}, fmt.Errorf("parse rss: %w", err)
	}

	out := ParsedFeed{Title: doc.Channel.Title}
	for _, it := range doc.Channel.Items {
		guid := it.GUID
		if guid == "" {
			guid = it.Link
		}

		item := ParsedItem{
			GUID:    guid,
			Title:   it.Title,
			Link:    it.Link,
			Summary: it.Description,
		}
		if t, err := parseRSSTime(it.PubDate); err == nil {
			item.PublishedAt = &t
		}
		out.Items = append(out.Items, item)
	}

	return out, nil
}

var rssTimeLayouts = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC3339,
	"02 Jan 2006 15:04:05 -0700",
	"2006-01-02 15:04:05",
}

func parseRSSTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty PubDate")
	}

	for _, layout := range rssTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized PubDate format: %s", s)
}
