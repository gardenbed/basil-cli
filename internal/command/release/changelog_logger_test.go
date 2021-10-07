package release

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name   string
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
			ui := cli.NewMockUi()
			l := newLogger(ui)

			l.ChangeVerbosity(0)

			l.Debug(tc.args...)
			l.Debugf(tc.format, tc.args...)

			l.Info(tc.args...)
			l.Infof(tc.format, tc.args...)

			l.Warn(tc.args...)
			l.Warnf(tc.format, tc.args...)

			l.Error(tc.args...)
			l.Errorf(tc.format, tc.args...)

			l.Fatal(tc.args...)
			l.Fatalf(tc.format, tc.args...)
		})
	}
}
