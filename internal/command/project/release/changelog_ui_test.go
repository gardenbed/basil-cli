package release

import (
	"testing"

	"github.com/gardenbed/charm/ui"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name   string
		style  ui.Style
		format string
		args   []interface{}
	}{
		{
			name:   "OK",
			format: "foo: %s",
			args:   []interface{}{"bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &changelogUI{
				UI: ui.NewNop(),
			}

			u.Tracef(tc.style, tc.format, tc.args...)
			u.Debugf(tc.style, tc.format, tc.args...)
			u.Infof(tc.style, tc.format, tc.args...)
			u.Warnf(tc.style, tc.format, tc.args...)
			u.Errorf(tc.style, tc.format, tc.args...)
		})
	}
}
