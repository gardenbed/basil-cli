package edit

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/debug"
)

func TestNewEditor(t *testing.T) {
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
			editor := NewEditor(tc.level)

			assert.NotNil(t, editor)
		})
	}
}

func TestEditor_Remove(t *testing.T) {
	assert.NoError(t, os.Mkdir("./temp", 0755))
	defer os.RemoveAll("./temp")

	tests := []struct {
		name          string
		globs         []string
		expectedError string
	}{
		{
			name:          "NoGlob",
			globs:         []string{},
			expectedError: "",
		},
		{
			name:          "NoMatch",
			globs:         []string{""},
			expectedError: "",
		},
		{
			name:          "InvalidPattern",
			globs:         []string{"\\"},
			expectedError: "syntax error in pattern",
		},
		{
			name:          "Success",
			globs:         []string{"./temp"},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			editor := &Editor{
				debugger: debug.NewSet(debug.None),
			}

			err := editor.Remove(tc.globs...)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestEditor_Move(t *testing.T) {
	err := os.Mkdir("./temp", 0755)
	assert.NoError(t, err)
	_, err = os.Create("./temp/foo")
	assert.NoError(t, err)
	defer os.RemoveAll("./temp")

	tests := []struct {
		name          string
		mkdir         bool
		specs         []MoveSpec
		expectedError string
	}{
		{
			name:          "NoSpec",
			mkdir:         false,
			specs:         []MoveSpec{},
			expectedError: "",
		},
		{
			name:  "InvalidSpec",
			mkdir: false,
			specs: []MoveSpec{
				{
					Src:  "./foo",
					Dest: "./bar",
				},
			},
			expectedError: "rename ./foo ./bar: no such file or directory",
		},
		{
			name:  "Success",
			mkdir: false,
			specs: []MoveSpec{
				{
					Src:  "./temp/foo",
					Dest: "./temp/bar",
				},
			},
			expectedError: "",
		},
		{
			name:  "Success_CreateDir",
			mkdir: true,
			specs: []MoveSpec{
				{
					Src:  "./temp/bar",
					Dest: "./temp/new/foo",
				},
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			editor := &Editor{
				debugger: debug.NewSet(debug.None),
			}

			err := editor.Move(tc.mkdir, tc.specs...)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestEditor_Append(t *testing.T) {
	err := os.Mkdir("./temp", 0755)
	assert.NoError(t, err)
	_, err = os.Create("./temp/foo")
	assert.NoError(t, err)
	defer os.RemoveAll("./temp")

	tests := []struct {
		name          string
		create        bool
		specs         []AppendSpec
		expectedError string
	}{
		{
			name:   "Success_Open",
			create: false,
			specs: []AppendSpec{
				{
					Path:    "./temp/foo",
					Content: "Hello, World!",
				},
			},
			expectedError: "",
		},
		{
			name:   "Success_Create",
			create: true,
			specs: []AppendSpec{
				{
					Path:    "./temp/bar",
					Content: "Hello, World!",
				},
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			editor := &Editor{
				debugger: debug.NewSet(debug.None),
			}

			err := editor.Append(tc.create, tc.specs...)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestEditor_ReplaceInDir(t *testing.T) {
	tests := []struct {
		name          string
		root          string
		specs         []ReplaceSpec
		expectedError string
	}{
		{
			name:          "DirectoryNotExist",
			root:          "./foo",
			specs:         []ReplaceSpec{},
			expectedError: "lstat ./foo: no such file or directory",
		},
		{
			name: "Success",
			root: "./test",
			specs: []ReplaceSpec{
				{
					PathRE: regexp.MustCompile(`\.txt$`),
					OldRE:  regexp.MustCompile(`foo`),
					New:    "bar",
				},
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			editor := &Editor{
				debugger: debug.NewSet(debug.None),
			}

			err := editor.ReplaceInDir(tc.root, tc.specs...)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
