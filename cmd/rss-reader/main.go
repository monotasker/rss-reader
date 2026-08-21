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

	feed, err := app.AddFeed("https://hnrss.org/frontpage")
	if err != nil {
		if err := app.RemoveFeed("https://hnrss.org/frontpage"); err != nil {
			return err
		}
		fmt.Println("removed feed")
	}
	fmt.Println("subscribed:", feed.Title, feed.ID)

	feeds, err := app.ListFeeds()
	for _, f := range feeds {
		fmt.Printf(" %s\n  %s\n", f.Title, f.URL)
	}

	return nil
}
