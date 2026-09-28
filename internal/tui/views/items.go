package views

import (
	"context"

	"github.com/monotasker/rss-reader/internal/service"
	"github.com/monotasker/rss-reader/internal/tui/shared"
)

type ItemsView struct {
	app   *service.App
	state *shared.AppState
}

func NewItemsView(a *service.App, s *shared.AppState) (*ItemsView, error) {
	newItemsView := &ItemsView{
		app:   a,
		state: s,
	}
	return newItemsView, nil
}

func (v *ItemsView) Name() string {
	return "ItemsView"
}

func (v *ItemsView) Render() error {
	return nil
}

func (v *ItemsView) HandleKey(ctx context.Context, key string) (ViewState, error) {
	switch key {
	case "j", "down":
		v.moveDown()
	case "k", "up":
		v.moveUp()
	case "r":
		v.refresh()
	case "q":
		return nil, nil
	case "f":
		return NewFeedsView(v.app, v.state)
	default:
		return v, nil
	}
	return v, nil
}

func (v *ItemsView) moveUp() error {
	return nil
}

func (v *ItemsView) moveDown() error {
	return nil
}

func (v *ItemsView) refresh() error {
	return nil
}
