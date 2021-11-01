package compile

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetModuleName(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedModule string
		expectedError  string
	}{
		{
			name:          "NoModFile",
			path:          "/opt",
			expectedError: "stat /opt/go.mod: no such file or directory",
		},
		{
			name:          "InvalidModule",
			path:          "./test/invalid_module",
			expectedError: "invalid go.mod file: no module name found",
		},
		{
			name:           "Success",
			path:           "./test/valid/lookup",
			expectedModule: "github.com/octocat/test",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			module, err := getModuleName(tc.path)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedModule, module)
			} else {
				assert.Empty(t, module)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestReadPackages(t *testing.T) {
	tests := []struct {
		name          string
		baseDir       string
		relDir        string
		visit         visitFunc
		expectedError string
	}{
		{
			name:    "InvalidPath",
			baseDir: "./dev/null",
			visit: func(_, _ string) error {
				return nil
			},
			expectedError: "open dev/null: no such file or directory",
		},
		{
			name:    "Success",
			baseDir: "./test/valid",
			visit: func(_, _ string) error {
				return nil
			},
			expectedError: "",
		},
		{
			name:    "VisitFails_FirstTime",
			baseDir: "./test/valid",
			visit: func(_, _ string) error {
				return errors.New("generic error")
			},
			expectedError: "generic error",
		},
		{
			name:    "VisitFails_SecondTime",
			baseDir: "./test/valid",
			visit: func(_, relDir string) error {
				if relDir == "." {
					return nil
				}
				return errors.New("generic error")
			},
			expectedError: "generic error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := readPackages(tc.baseDir, tc.visit)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
