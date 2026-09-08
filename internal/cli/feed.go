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
}
