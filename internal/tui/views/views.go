/*
	Package views

TUI views for rss-reader.
*/
package views

import "context"

type ViewState interface {
	Render() error
	HandleKey(ctx context.Context, key string) (ViewState, error)
	Name() string
}
