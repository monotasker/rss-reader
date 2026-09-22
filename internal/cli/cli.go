package cli

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/monotasker/rss-reader/internal/service"
)

// commandFunc is the signature every subcommand implements
type commandFunc func(ctx context.Context, app *service.App, args []string) error

// command pairs an implementation with its help text
type command struct {
	run     commandFunc
	summary string
}

// commands maps the first CLI argument to its implementation
var commands = map[string]command{
	"feed": {runFeed, "manage feed subscriptions (add, list, remove)"},
}

// usageError marks "the user typed it wrong" failures so Run can exit 2
// and show usage help text, instead of exit 1
type usageError struct{ msg string }

func (e usageError) Error() string { return e.msg }

func usagef(format string, args ...any) error {
	return usageError{fmt.Sprintf(format, args...)}
}

// Run dispatches args (os.Args[1:]) and returns the process exit code.
func Run(app *service.App, args []string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		printUsage()
		return 0
	}

	ctx := context.Background()

	cmd, ok := commands[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		printUsage()
		return 2
	}

	err := cmd.run(ctx, app, args[1:])
	var ue usageError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &ue):
		fmt.Fprintln(os.Stderr, "usage error:", ue.msg)
		return 2
	default:
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
}

func printUsage() {
	fmt.Println("rss-reader - a fast, simple feed reader")
	fmt.Println("Commands:")

	keys := make([]string, 0, len(commands))
	for key := range commands {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, cmp.Compare)

	for _, k := range keys {
		fmt.Printf("    %-10s %s\n", k, commands[k].summary)
	}
}
