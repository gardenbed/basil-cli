package ui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/moorara/promptui"
	"github.com/moorara/promptui/list"
)

const detailsTemplate = `{{ if .Attributes }}
-------------------- Details --------------------
{{ range $i, $a := .Attributes }}{{ $a.Key }}: {{ $a.Value | faint }}
{{ end }}{{ end }}`

// interactiveUI implements the UI interface.
type interactiveUI struct {
	sync.Mutex
	level       Level
	reader      io.ReadCloser
	writer      io.WriteCloser
	errorWriter io.WriteCloser
}

// NewInteractive creates a new interactive user interface.
// This is a concurrent-safe UI and can be used across multiple Go routines.
func NewInteractive(level Level) UI {
	return &interactiveUI{
		level:       level,
		reader:      os.Stdin,
		writer:      os.Stdout,
		errorWriter: os.Stderr,
	}
}

func (u *interactiveUI) Printf(format string, a ...interface{}) {
	u.Lock()
	defer u.Unlock()

	s := fmt.Sprintf(format, a...)
	fmt.Fprintln(u.writer, s)
}

func (u *interactiveUI) GetLevel() Level {
	u.Lock()
	defer u.Unlock()

	return u.level
}

func (u *interactiveUI) SetLevel(l Level) {
	u.Lock()
	defer u.Unlock()

	u.level = l
}

func (u *interactiveUI) Tracef(style Style, format string, a ...interface{}) {
	u.Lock()
	defer u.Unlock()

	if u.level <= Trace {
		s := style.sprintf(format, a...)
		fmt.Fprintln(u.writer, s)
	}
}

func (u *interactiveUI) Debugf(style Style, format string, a ...interface{}) {
	u.Lock()
	defer u.Unlock()

	if u.level <= Debug {
		s := style.sprintf(format, a...)
		fmt.Fprintln(u.writer, s)
	}
}

func (u *interactiveUI) Infof(style Style, format string, a ...interface{}) {
	u.Lock()
	defer u.Unlock()

	if u.level <= Info {
		s := style.sprintf(format, a...)
		fmt.Fprintln(u.writer, s)
	}
}

func (u *interactiveUI) Warnf(style Style, format string, a ...interface{}) {
	u.Lock()
	defer u.Unlock()

	if u.level <= Warn {
		s := style.sprintf(format, a...)
		fmt.Fprintln(u.writer, s)
	}
}

func (u *interactiveUI) Errorf(style Style, format string, a ...interface{}) {
	u.Lock()
	defer u.Unlock()

	if u.level <= Error {
		s := style.sprintf(format, a...)
		fmt.Fprintln(u.errorWriter, s)
	}
}

func (u *interactiveUI) Confrim(prompt string, Default bool) (bool, error) {
	u.Lock()
	defer u.Unlock()

	p := promptui.Prompt{
		Label:     prompt,
		IsConfirm: true,
		Stdin:     u.reader,
		Stdout:    u.writer,
	}

	if Default {
		p.Default = "Y"
	}

	if _, err := p.Run(); err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (u *interactiveUI) Ask(prompt, Default string, validate ValidateFunc) (string, error) {
	u.Lock()
	defer u.Unlock()

	p := promptui.Prompt{
		Label:    prompt,
		Default:  Default,
		Validate: promptui.ValidateFunc(validate),
		Stdin:    u.reader,
		Stdout:   u.writer,
	}

	return p.Run()
}

func (u *interactiveUI) AskSecret(prompt string, confirm bool, validate ValidateFunc) (string, error) {
	u.Lock()
	defer u.Unlock()

	p1 := promptui.Prompt{
		Label:    prompt,
		Mask:     '•',
		Validate: promptui.ValidateFunc(validate),
		Stdin:    u.reader,
		Stdout:   u.writer,
	}

	first, err := p1.Run()
	if err != nil {
		return "", err
	}

	if !confirm {
		return first, nil
	}

	// Create a new prompt to avoid race conditions
	p2 := promptui.Prompt{
		Label:    fmt.Sprintf("%s (confirmation)", prompt),
		Mask:     '•',
		Validate: promptui.ValidateFunc(validate),
		Stdin:    u.reader,
		Stdout:   u.writer,
	}

	// Confirm the input
	second, err := p2.Run()
	if err != nil {
		return "", err
	}

	if first != second {
		return "", errors.New("confirmation not matching")
	}

	return second, nil
}

func (u *interactiveUI) Select(prompt string, size int, items []Item, search SearchFunc) (Item, error) {
	u.Lock()
	defer u.Unlock()

	templates := &promptui.SelectTemplates{
		// Label: "{{ . }}?",
		Active:   `{{ "➜" | yellow }} {{ .Name | cyan }} {{ printf "(%s)" .Description | faint }}`,
		Inactive: `  {{ .Name | blue }}`,
		Selected: `{{ "✓" | green }} {{ .Name | faint }}`,
		Details:  detailsTemplate,
	}

	p := promptui.Select{
		Label:     prompt,
		Items:     items,
		Size:      size,
		Templates: templates,
		Searcher:  list.Searcher(search),
	}

	i, _, err := p.Run()
	if err != nil {
		return Item{}, err
	}

	return items[i], nil
}
