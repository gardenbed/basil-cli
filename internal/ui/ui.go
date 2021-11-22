package ui

import "github.com/gardenbed/charm/ui"

var (
	Trace = ui.Trace
	Debug = ui.Debug
	Info  = ui.Info
	Warn  = ui.Warn
	Error = ui.Error
	None  = ui.None
)

var (
	Magenta = ui.Magenta
	Cyan    = ui.Green
	Green   = ui.Green
	Yellow  = ui.Yellow
	Red     = ui.Red
)

type (
	Attribute struct {
		Key   string
		Value string
	}

	// Item represents an item for selection.
	Item struct {
		Key         string
		Name        string
		Description string
		Attributes  []Attribute
	}
)

type (
	// ValidateFunc is a function for validating user inputs.
	ValidateFunc func(string) error

	// SearchFunc is a function for searching and matching items in select.
	SearchFunc func(string, int) bool
)

// UI is the interface for interacting with users in command-line applications.
type UI interface {
	ui.UI
	Confrim(prompt string, Default bool) (bool, error)
	Ask(prompt, Default string, f ValidateFunc) (string, error)
	AskSecret(prompt string, confirm bool, f ValidateFunc) (string, error)
	Select(prompt string, size int, items []Item, f SearchFunc) (Item, error)
}
