package archive

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/debug"
)

func TestNewTarArchive(t *testing.T) {
	tests := []struct {
		name  string
		level debug.Level
	}{
		{
			name:  "OK",
			level: debug.None,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			arch := NewTarArchive(tc.level)

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
			dest, err := ioutil.TempDir("", "gelato-test-*")
			assert.NoError(t, err)
			defer os.RemoveAll(dest)

			f, err := os.Open(tc.archFile)
			assert.NoError(t, err)

			arch := &TarArchive{
				debugger: debug.NewSet(debug.None),
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