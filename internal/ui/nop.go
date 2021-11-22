package ui

import "github.com/gardenbed/charm/ui"

type nopUI struct {
	ui.UI
}

// NewNop creates a nop user interface for testing purposes.
func NewNop() UI {
	return &nopUI{
		UI: ui.NewNop(),
	}
}

func (u *nopUI) Confrim(string, bool) (bool, error) {
	return true, nil
}

func (u *nopUI) Ask(string, string, ValidateFunc) (string, error) {
	return "", nil
}

func (u *nopUI) AskSecret(string, bool, ValidateFunc) (string, error) {
	return "", nil
}

func (u *nopUI) Select(string, int, []Item, SearchFunc) (Item, error) {
	return Item{}, nil
}
