package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStyle_sprintf(t *testing.T) {
	tests := []struct {
		name           string
		s              Style
		format         string
		args           []interface{}
		expectedString string
	}{
		{
			name:           "Bold",
			s:              Style{Bold},
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedString: "\x1b[1mHello, World!\x1b[0m",
		},
		{
			name:           "FgGreen",
			s:              Style{FgGreen},
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedString: "\x1b[32mHello, World!\x1b[0m",
		},
		{
			name:           "BgBlue",
			s:              Style{BgBlue},
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedString: "\x1b[44mHello, World!\x1b[0m",
		},
		{
			name:           "MixStyle",
			s:              Style{BgYellow, FgMagenta, Bold, BlinkSlow},
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedString: "\x1b[43;35;1;5mHello, World!\x1b[0m",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.s.sprintf(tc.format, tc.args...)

			assert.Equal(t, tc.expectedString, s)
		})
	}
}
