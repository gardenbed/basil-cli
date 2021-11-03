// Package debug is used for printing debugging information and metadata.
package debug

import (
	"log"
	"os"
	"sync"

	"github.com/fatih/color"
)

// Level is the verbosity level.
type Level int

const (
	// Trace shows information in all levels.
	Trace Level = iota
	// Debug shows information in Debug, Info, Warn, and Error levels.
	Debug
	// Info shows information in Info, Warn, and Error levels.
	Info
	// Warn shows information in Warn and Error levels.
	Warn
	// Error shows information only in Error leved.
	Error
	// None does not show any information.
	None
)

// Debugger is the interface for printing debugging information.
type Debugger interface {
	Level() Level
	Tracef(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

// debugger implements the Debugger interface.
type debugger struct {
	sync.Mutex
	level  Level
	logger *log.Logger
}

// New creates a new debugger.
func New(level Level) Debugger {
	return &debugger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

func (d *debugger) Level() Level {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	return d.level
}

func (d *debugger) Tracef(format string, v ...interface{}) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	if d.level <= Trace {
		d.logger.Printf(format, v...)
	}
}

func (d *debugger) Debugf(format string, v ...interface{}) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	if d.level <= Debug {
		d.logger.Printf(format, v...)
	}
}

func (d *debugger) Infof(format string, v ...interface{}) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	if d.level <= Info {
		d.logger.Printf(format, v...)
	}
}

func (d *debugger) Warnf(format string, v ...interface{}) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	if d.level <= Warn {
		d.logger.Printf(format, v...)
	}
}

func (d *debugger) Errorf(format string, v ...interface{}) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	if d.level <= Error {
		d.logger.Printf(format, v...)
	}
}

type coloredDebugger struct {
	debugger *debugger
	color    *color.Color
}

// NewColored creates a new colored debugger.
func NewColored(level Level, color *color.Color) Debugger {
	debugger := &debugger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}

	return &coloredDebugger{
		debugger: debugger,
		color:    color,
	}
}

func (d *coloredDebugger) Level() Level {
	return d.debugger.Level()
}

func (d *coloredDebugger) Tracef(format string, v ...interface{}) {
	msg := d.color.Sprintf(format, v...)
	d.debugger.Tracef(msg)
}

func (d *coloredDebugger) Debugf(format string, v ...interface{}) {
	msg := d.color.Sprintf(format, v...)
	d.debugger.Debugf(msg)
}

func (d *coloredDebugger) Infof(format string, v ...interface{}) {
	msg := d.color.Sprintf(format, v...)
	d.debugger.Infof(msg)
}

func (d *coloredDebugger) Warnf(format string, v ...interface{}) {
	msg := d.color.Sprintf(format, v...)
	d.debugger.Warnf(msg)
}

func (d *coloredDebugger) Errorf(format string, v ...interface{}) {
	msg := d.color.Sprintf(format, v...)
	d.debugger.Errorf(msg)
}

// DebuggerSet is a collection of colored debuggers.
type DebuggerSet struct {
	Red, Green, Yellow, Blue, Magenta, Cyan, White Debugger
}

// NewSet creates a new set of colored debuggers.
func NewSet(level Level) *DebuggerSet {
	debugger := &debugger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}

	return &DebuggerSet{
		Red: &coloredDebugger{
			debugger: debugger,
			color:    color.New(color.FgRed),
		},
		Green: &coloredDebugger{
			debugger: debugger,
			color:    color.New(color.FgGreen),
		},
		Yellow: &coloredDebugger{
			debugger: debugger,
			color:    color.New(color.FgYellow),
		},
		Blue: &coloredDebugger{
			debugger: debugger,
			color:    color.New(color.FgBlue),
		},
		Magenta: &coloredDebugger{
			debugger: debugger,
			color:    color.New(color.FgMagenta),
		},
		Cyan: &coloredDebugger{
			debugger: debugger,
			color:    color.New(color.FgCyan),
		},
		White: &coloredDebugger{
			debugger: debugger,
			color:    color.New(color.FgWhite),
		},
	}
}
