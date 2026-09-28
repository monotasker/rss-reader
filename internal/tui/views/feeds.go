package views

import (
	"context"

	"github.com/monotasker/rss-reader/internal/service"
	"github.com/monotasker/rss-reader/internal/tui/shared"
)

type FeedsView struct {
	app   *service.App
	state *shared.AppState
}

func NewFeedsView(a *service.App, s *shared.AppState) (*FeedsView, error) {
	newFeedsView := &FeedsView{
		app:   a,
		state: s,
	}
	return newFeedsView, nil
}

func (v *FeedsView) Name() string {
	return "FeedsView"
}

func (v *FeedsView) Render() error {
	return nil
}

func (v *FeedsView) HandleKey(ctx context.Context, key string) (ViewState, error) {
	switch key {
	case "j", "down":
		v.moveDown()
	case "k", "up":
		v.moveUp()
	case "r":
		v.refresh()
	case "q":
		return nil, nil
	case "escape":
		return NewItemsView(v.app, v.state)
	default:
		return v, nil
	}
	return v, nil
}

func (v *FeedsView) moveUp() error {
	return nil
}

func (v *FeedsView) moveDown() error {
	return nil
}

func (v *FeedsView) refresh() error {
	return nil
}
