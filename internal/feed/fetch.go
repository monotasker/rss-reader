package feed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetcher downloads documents with a shared, pooled HTTP client.
type Fetcher struct {
	client *http.Client
	config *Config
}

type Config struct {
	Timeout      time.Duration
	UserAgent    string
	MaxFeedBytes int
}

var defaultConfig = Config{
	Timeout:      30 * time.Second,
	UserAgent:    "rss-reader/0.2 (+https://github.com/monotasker/rss-reader)",
	MaxFeedBytes: 10 << 20, // 10 MiB
}

type Option func(*Config)

func WithTimeout(d time.Duration) Option {
	return func(c *Config) { c.Timeout = d }
}

func WithUserAgent(u string) Option {
	return func(c *Config) { c.UserAgent = u }
}

func WithMaxFeedBytes(b int) Option {
	return func(c *Config) { c.MaxFeedBytes = b }
}

// NewFetcher returns a Fetcher with sane defaults
func NewFetcher(options ...Option) (*Fetcher, error) {
	config := &defaultConfig
	for _, opt := range options {
		opt(config)
	}
	client := &http.Client{Timeout: config.Timeout}
	return &Fetcher{
		client: client,
		config: config,
	}, nil
}

// Fetch downloads the raw feed document.
// If successful, it returns the response as bytes, usually
// in XML format. For non-2xx responses it returns an error.
func (fetcher *Fetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Fetch %s: %w", url, err)
	}
	request.Header.Set("User-Agent", fetcher.config.UserAgent)

	response, err := fetcher.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Fetch %s: %w", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Fetch %s: unexpected status %s", url, response.Status)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, int64(fetcher.config.MaxFeedBytes+1)))
	if err != nil {
		return nil, fmt.Errorf("Fetch %s: read body %w", url, err)
	}
	if len(body) > fetcher.config.MaxFeedBytes {
		return nil, fmt.Errorf("Fetch %s: response body exceeds %d byte limit", url, fetcher.config.MaxFeedBytes)
	}
	return body, nil
}
