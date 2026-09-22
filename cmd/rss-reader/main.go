package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/monotasker/rss-reader/internal/cli"
	"github.com/monotasker/rss-reader/internal/service"
	"github.com/monotasker/rss-reader/internal/store"
)

const Version = "0.2.0-dev"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	st, err := store.Open("rss-reader.db")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error opening store:", err)
		os.Exit(1)
	}
	defer st.Close()

	app, err := service.New(st, logger)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating application:", err)
	}
	code := cli.Run(app, os.Args[1:])
	st.Close()
	os.Exit(code)
}
