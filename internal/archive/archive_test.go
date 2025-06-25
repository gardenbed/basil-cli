package archive

import (
	"os"
	"testing"

	"github.com/gardenbed/charm/ui"
	"github.com/stretchr/testify/assert"
)

func TestNewTarArchive(t *testing.T) {
	tests := []struct {
		name string
		ui   ui.UI
	}{
		{
			name: "OK",
			ui:   ui.NewNop(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			arch := NewTarArchive(tc.ui)

			assert.NotNil(t, arch)
		})
	}
}

func TestTarArchive_Extract(t *testing.T) {
	tests := []struct {
		name          string
		archFile      string
		f             Selector
		expectedError string
	}{
		{
			name:          "InvalidArchive",
			archFile:      "test/invalid.tar.gz",
			f:             nil,
			expectedError: "error on creating gzip reader: EOF",
		},
		{
			name:     "Success",
			archFile: "test/github.tar.gz",
			f: func(path string) (string, bool) {
				return path, true
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dest, err := os.MkdirTemp("", "gelato-test-*")
			assert.NoError(t, err)

			defer func() {
				assert.NoError(t, os.RemoveAll(dest))
			}()

			f, err := os.Open(tc.archFile)
			assert.NoError(t, err)

			arch := &TarArchive{
				ui: ui.NewNop(),
			}

			err = arch.Extract(dest, f, tc.f)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
