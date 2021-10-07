package release

import (
	"sync"

	"github.com/fatih/color"
	"github.com/gardenbed/changelog/log"
	"github.com/mitchellh/cli"
)

const indent = "  "

// logger implements the github.com/gardenbed/changelog/log.Logger interface
type logger struct {
	sync.Mutex
	ui     cli.Ui
	colors struct {
		info  *color.Color
		warn  *color.Color
		err   *color.Color
		fatal *color.Color
	}
}

func newLogger(ui cli.Ui) *logger {
	l := &logger{
		ui: ui,
	}

	l.colors.info = color.New(color.FgCyan)
	l.colors.warn = color.New(color.FgYellow)
	l.colors.err = color.New(color.FgRed)
	l.colors.fatal = color.New(color.FgRed)

	return l
}

func (l *logger) ChangeVerbosity(v log.Verbosity)        {}
func (l *logger) Debug(v ...interface{})                 {}
func (l *logger) Debugf(format string, v ...interface{}) {}

func (l *logger) Info(v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.ui.Output(indent + l.colors.info.Sprint(v...))
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.ui.Output(indent + l.colors.info.Sprintf(format, v...))
}

func (l *logger) Warn(v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.ui.Output(indent + l.colors.warn.Sprint(v...))
}

func (l *logger) Warnf(format string, v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.ui.Output(indent + l.colors.warn.Sprintf(format, v...))
}

func (l *logger) Error(v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.ui.Output(indent + l.colors.err.Sprint(v...))
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.ui.Output(indent + l.colors.err.Sprintf(format, v...))
}

func (l *logger) Fatal(v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.ui.Output(indent + l.colors.fatal.Sprint(v...))
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.ui.Output(indent + l.colors.fatal.Sprintf(format, v...))
}
