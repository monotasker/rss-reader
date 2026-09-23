package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/monotasker/rss-reader/internal/service"
	"github.com/monotasker/rss-reader/internal/store"
	"github.com/monotasker/rss-reader/internal/utils"
)

func runFeed(ctx context.Context, app *service.App, args []string) error {
	if len(args) == 0 {
		return usagef("feed requires a subcommand: add, list, remove")
	}

	switch args[0] {
	case "add":
		return feedAdd(ctx, app, args[1:])
	case "list":
		return feedList(ctx, app, args[1:])
	case "refresh":
		return feedRefresh(ctx, app, args[1:])
	case "remove":
		return feedRemove(ctx, app, args[1:])
	default:
		return usagef("unknown feed subcommand %q", args[0])
	}
}

func feedAdd(ctx context.Context, app *service.App, args []string) error {
	if len(args) != 1 {
		return usagef("feed add <url>")
	}
	f, err := app.AddFeed(ctx, args[0])
	if errors.Is(err, store.ErrDuplicate) {
		return fmt.Errorf("already subsccribed to %s", args[0])
	}
	if err != nil {
		return err
	}
	fmt.Printf("subscribed to %s (%s)\n", f.Title, f.ID)
	return nil
}

func feedList(ctx context.Context, app *service.App, args []string) error {
	fs := flag.NewFlagSet("feed list", flag.ContinueOnError)
	long := fs.Bool("long", false, "include id and url columns")
	if err := fs.Parse(args); err != nil {
		return usagef("feed list: %v", err)
	}

	feeds, err := app.ListFeeds(ctx)
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
		fmt.Fprintln(w, "TITLE\tID\tURL\tLAST FETCHED")
		for _, f := range feeds {
			readableFetched := "never"
			if f.LastFetched != nil {
				relativeFetched := utils.RelativeTime(*f.LastFetched)
				readableFetched = utils.HumanReadableDuration(relativeFetched)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", f.Title, f.ID, f.URL, readableFetched)
		}
		return nil
	}
	fmt.Fprintln(w, "Title\tURL")
	for _, f := range feeds {
		fmt.Fprintf(w, "%s\t%s\n", f.Title, f.URL)
	}
	return nil
}

func feedRefresh(ctx context.Context, app *service.App, args []string) error {
	if len(args) != 1 {
		return usagef("feed refresh <url-or-id>")
	}
	result, err := app.RefreshFeed(ctx, args[0])
	if err != nil {
		return err
	}
	fmt.Printf("%s: %d new items\n", result.Feed.Title, result.NewItems)
	return nil
}

func feedRemove(ctx context.Context, app *service.App, args []string) error {
	if len(args) != 1 {
		return usagef("feed remove <url-or-id>")
	}
	f, err := app.RemoveFeed(ctx, args[0])
	if err != nil {
		return err
	}
	fmt.Printf("removed: %s\n", f.Title)
	return nil
}
