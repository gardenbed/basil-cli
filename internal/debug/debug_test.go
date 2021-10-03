package debug

import (
	"testing"

	"github.com/fatih/color"

	"github.com/stretchr/testify/assert"
)

func TestDebugger(t *testing.T) {
	tests := []struct {
		name   string
		level  Level
		format string
		args   []interface{}
	}{
		{
			name:   "Trace",
			level:  Trace,
			format: "foo: %s",
			args:   []interface{}{"bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := New(tc.level)

			assert.Equal(t, tc.level, d.Level())

			d.Tracef(tc.format, tc.args...)
			d.Debugf(tc.format, tc.args...)
			d.Infof(tc.format, tc.args...)
			d.Warnf(tc.format, tc.args...)
			d.Errorf(tc.format, tc.args...)
		})
	}
}

func TestColoredDebugger(t *testing.T) {
	tests := []struct {
		name   string
		level  Level
		color  *color.Color
		format string
		args   []interface{}
	}{
		{
			name:   "Trace",
			level:  Trace,
			color:  color.New(color.FgWhite),
			format: "foo: %s",
			args:   []interface{}{"bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewColored(tc.level, tc.color)

			assert.Equal(t, tc.level, d.Level())

			d.Tracef(tc.format, tc.args...)
			d.Debugf(tc.format, tc.args...)
			d.Infof(tc.format, tc.args...)
			d.Warnf(tc.format, tc.args...)
			d.Errorf(tc.format, tc.args...)
		})
	}
}

func TestDebuggerSet(t *testing.T) {
	tests := []struct {
		name   string
		level  Level
		format string
		args   []interface{}
	}{
		{
			name:   "Trace",
			level:  Trace,
			format: "foo: %s",
			args:   []interface{}{"bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewSet(tc.level)

			d.Red.Tracef(tc.format, tc.args...)
			d.Red.Debugf(tc.format, tc.args...)
			d.Red.Infof(tc.format, tc.args...)
			d.Red.Warnf(tc.format, tc.args...)
			d.Red.Errorf(tc.format, tc.args...)

			d.Green.Tracef(tc.format, tc.args...)
			d.Green.Debugf(tc.format, tc.args...)
			d.Green.Infof(tc.format, tc.args...)
			d.Green.Warnf(tc.format, tc.args...)
			d.Green.Errorf(tc.format, tc.args...)

			d.Yellow.Tracef(tc.format, tc.args...)
			d.Yellow.Debugf(tc.format, tc.args...)
			d.Yellow.Infof(tc.format, tc.args...)
			d.Yellow.Warnf(tc.format, tc.args...)
			d.Yellow.Errorf(tc.format, tc.args...)

			d.Blue.Tracef(tc.format, tc.args...)
			d.Blue.Debugf(tc.format, tc.args...)
			d.Blue.Infof(tc.format, tc.args...)
			d.Blue.Warnf(tc.format, tc.args...)
			d.Blue.Errorf(tc.format, tc.args...)

			d.Magenta.Tracef(tc.format, tc.args...)
			d.Magenta.Debugf(tc.format, tc.args...)
			d.Magenta.Infof(tc.format, tc.args...)
			d.Magenta.Warnf(tc.format, tc.args...)
			d.Magenta.Errorf(tc.format, tc.args...)

			d.Cyan.Tracef(tc.format, tc.args...)
			d.Cyan.Debugf(tc.format, tc.args...)
			d.Cyan.Infof(tc.format, tc.args...)
			d.Cyan.Warnf(tc.format, tc.args...)
			d.Cyan.Errorf(tc.format, tc.args...)

			d.White.Tracef(tc.format, tc.args...)
			d.White.Debugf(tc.format, tc.args...)
			d.White.Infof(tc.format, tc.args...)
			d.White.Warnf(tc.format, tc.args...)
			d.White.Errorf(tc.format, tc.args...)
		})
	}
}
