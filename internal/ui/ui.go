package ui

// Level represents the verbosity level.
type Level int

const (
	// Trace shows all messages.
	Trace Level = iota
	// Debug shows Debug, Info, Warn, and Error messages.
	Debug
	// Info shows Info, Warn, and Error messages.
	Info
	// Warn shows Warn and Error messages.
	Warn
	// Error shows only Error messages.
	Error
	// None does not show any messages.
	None
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
	// Output method independent of the verbosity level
	Printf(format string, a ...interface{})

	// Leveled output methods
	GetLevel() Level
	SetLevel(l Level)
	Tracef(s Style, format string, a ...interface{})
	Debugf(s Style, format string, a ...interface{})
	Infof(s Style, format string, a ...interface{})
	Warnf(s Style, format string, a ...interface{})
	Errorf(s Style, format string, a ...interface{})

	// User input methods
	Confrim(prompt string, Default bool) (bool, error)
	Ask(prompt, Default string, f ValidateFunc) (string, error)
	AskSecret(prompt string, confirm bool, f ValidateFunc) (string, error)
	Select(prompt string, size int, items []Item, f SearchFunc) (Item, error)
}
