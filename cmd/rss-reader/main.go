package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/monotasker/rss-reader/internal/service"
	"github.com/monotasker/rss-reader/internal/store"
)

const Version = "0.2.0-dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	logger.Info("rss-reader starting", "version", Version)

	st, err := store.Open("rss-reader.db")
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer st.Close()

	app := service.New(st, logger)

	parsed, err := app.PreviewFeed("https://hnrss.org/frontpage")
	if err != nil {
		return err
	}

	fmt.Printf("%s -- %d items\n", parsed.Title, len(parsed.Items))
	for i, item := range parsed.Items {
		if i >= 5 {
			break
		}
		fmt.Printf(" %s\n  %s\n", item.Title, item.Link)
	}

	return nil
}
