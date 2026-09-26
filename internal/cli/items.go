package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/monotasker/rss-reader/internal/domain"
	"github.com/monotasker/rss-reader/internal/service"
	"github.com/monotasker/rss-reader/internal/store"
)

func runItems(ctx context.Context, app *service.App, args []string) error {
	if len(args) == 0 {
		return usagef("items requires a subcommand: list, mark")
	}
	switch args[0] {
	case "list":
		return itemsList(ctx, app, args[1:])
	case "mark":
		return itemsMark(ctx, app, args[1:])
	default:
		return usagef("unknown items subcommand %q", args[0])
	}
}

func itemsList(ctx context.Context, app *service.App, args []string) error {
	flagSet := flag.NewFlagSet("items list", flag.ContinueOnError)
	feedRef := flagSet.String("feed", "", "only items from this feed (url or id)")
	all := flagSet.Bool("all", false, "include items already read")
	bookmarked := flagSet.Bool("bookmarked", false, "only bookmarked items")
	dismissed := flagSet.Bool("dismissed", false, "include dismissed items")
	search := flagSet.String("search", "", "match text in title or summary")
	pageSize := flagSet.Int("n", 20, "items per page")
	page := flagSet.Int("page", 1, "page number")
	if err := flagSet.Parse(args); err != nil {
		return usagef("items list: %v", err)
	}

	items, err := app.ListItems(ctx, store.ItemFilter{
		FeedID:           *feedRef,
		UnreadOnly:       !*all && !*bookmarked,
		BookmarkedOnly:   *bookmarked,
		IncludeDismissed: *dismissed,
		Search:           strings.TrimSpace(*search),
		Limit:            *pageSize,
		Offset:           (*page - 1) * *pageSize,
	})
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Println("no items — try: `rss-reader feed refresh` or `items list -all`")
		return nil
	}

	feeds, err := app.ListFeeds(ctx)
	if err != nil {
		return err
	}
	feedTitles := make(map[string]string, len(feeds))
	for _, feed := range feeds {
		feedTitles[feed.ID] = feed.Title
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	defer w.Flush()
	fmt.Fprintln(w, "ID\t\tDATE\tFEED\tTITLE")
	for _, item := range items {
		marker := " "
		if !item.Read {
			marker = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID[28:36], marker,
			item.PublishedAt.Format("2006-01-02 15:04"),
			feedTitles[item.FeedID],
			strings.TrimSpace(item.Title))
	}
	return nil
}

func itemsMark(ctx context.Context, app *service.App, args []string) error {
	if len(args) < 2 {
		return usagef("items mark <read|unread|bookmarked|unbookmarked|dismissed|undismissed> <id>...")
	}
	verb, refs := args[0], args[1:]
	for _, ref := range refs {
		item, err := markOne(ctx, app, verb, ref)
		if errors.Is(err, store.ErrAmbiguous) {
			return fmt.Errorf("id %q matches more than one item — give more characters", ref)
		}
		if err != nil {
			return err
		}
		fmt.Printf("%s: %s  %s\n", verb, item.ID[28:36],
			strings.TrimSpace(item.Title))
	}
	return nil
}

func markOne(ctx context.Context, app *service.App, verb, ref string) (domain.Item, error) {
	switch verb {
	case "read":
		return app.MarkItemRead(ctx, ref)
	case "unread":
		return app.MarkItemUnread(ctx, ref)
	case "bookmarked":
		return app.BookmarkItem(ctx, ref)
	case "unbookmarked":
		return app.UnbookmarkItem(ctx, ref)
	case "dismissed":
		return app.DismissItem(ctx, ref)
	case "undismissed":
		return app.UndismissItem(ctx, ref)
	default:
		return domain.Item{}, usagef("items mark: unknown state %q", verb)
	}
}
