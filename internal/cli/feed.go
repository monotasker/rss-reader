package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/monotasker/rss-reader/internal/service"
	"github.com/monotasker/rss-reader/internal/store"
)

func runFeed(app *service.App, args []string) error {
	if len(args) == 0 {
		return usagef("feed requires a subcommand: add, list, remove")
	}

	switch args[0] {
	case "add":
		return feedAdd(app, args[1:])
	case "list":
		return feedList(app, args[1:])
	case "remove":
		return feedRemove(app, args[1:])
	default:
		return usagef("unknown feed subcommand %q", args[0])
	}
}

func feedAdd(app *service.App, args []string) error {
	if len(args) != 1 {
		return usagef("feed add <url>")
	}
	f, err := app.AddFeed(args[0])
	if errors.Is(err, store.ErrDuplicate) {
		return fmt.Errorf("already subsccribed to %s", args[0])
	}
	if err != nil {
		return err
	}
	fmt.Printf("subscribed to %s (%s)\n", f.Title, f.ID)
	return nil
}

func feedList(app *service.App, args []string) error {
	fs := flag.NewFlagSet("feed list", flag.ContinueOnError)
	long := fs.Bool("long", false, "include id and url columns")
	if err := fs.Parse(args); err != nil {
		return usagef("feed list: %v", err)
	}

	feeds, err := app.ListFeeds()
	if err != nil {
		return err
	}
	if len(feeds) == 0 {
		fmt.Println("no feeds subscribed - try: rss-reader feed add <url>")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	defer w.Flush()
	if *long {
		fmt.Fprintln(w, "TITLE\tID\tURL")
		for _, f := range feeds {
			fmt.Fprintf(w, "%s\t%s\t%s\n", f.Title, f.ID, f.URL)
		}
		return nil
	}
	fmt.Fprintln(w, "Title\tURL")
	for _, f := range feeds {
		fmt.Fprintf(w, "%s\t%s\n", f.Title, f.URL)
	}
	return nil
}

func feedRemove(app *service.App, args []string) error {
	if len(args) != 1 {
		return usagef("feed remove <url-or-id>")
	}
	f, err := app.RemoveFeed(args[0])
	if err != nil {
		return err
	}
	fmt.Printf("removed: %s\n", f.Title)
	return nil
}
