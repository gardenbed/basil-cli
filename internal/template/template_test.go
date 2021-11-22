package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gardenbed/charm/ui"
	"github.com/stretchr/testify/assert"
)

func prepare(t *testing.T, path string) func() {
	tempPath := filepath.Join(path, "./temp")

	assert.NoError(t,
		os.Mkdir(tempPath, 0755),
	)

	assert.NoError(t,
		os.WriteFile(filepath.Join(tempPath, "foo"), []byte("Lorem ipsum\n"), 0644),
	)

	assert.NoError(t,
		os.WriteFile(filepath.Join(tempPath, "bar"), []byte("Lorem ipsum\n"), 0644),
	)

	return func() {
		_ = os.RemoveAll(tempPath)
	}
}

func TestRead(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		params           interface{}
		expectedTemplate Template
		expectedError    string
	}{
		{
			name:          "NoFile",
			path:          "./test",
			params:        nil,
			expectedError: "template file not found",
		},
		{
			name:          "EmptyYAML",
			path:          "./test/empty",
			params:        nil,
			expectedError: "EOF",
		},
		{
			name:          "InvalidYAML",
			path:          "./test/invalid-yaml",
			params:        nil,
			expectedError: "yaml: unmarshal errors",
		},
		{
			name:          "InvalidTemplate",
			path:          "./test/invalid-template",
			params:        nil,
			expectedError: "unterminated character constant",
		},
		{
			name:          "InvalidField",
			path:          "./test/invalid-field",
			params:        struct{}{},
			expectedError: "can't evaluate field Unknown",
		},
		{
			name: "Success",
			path: "./test/valid",
			params: struct {
				Name string
			}{
				Name: "placereleaser",
			},
			expectedTemplate: Template{
				path:        "./test/valid",
				Name:        "test-template",
				Description: "This template is used for testing.",
				Changes: Changes{
					Deletes: Deletes{
						{Glob: "template.yaml"},
						{Glob: ".git"},
						{Glob: "*.pb.go"},
					},
					Moves: Moves{
						{
							Src:  "./cmd/placeholder",
							Dest: "./cmd/placereleaser",
						},
						{
							Src:  "./.github/workflows/placeholder.yml",
							Dest: "./.github/workflows/placereleaser.yml",
						},
					},
					Appends: Appends{
						{
							Filepath: "./.github/CODEOWNERS",
							Content:  "@octocat",
						},
					},
					Replaces: Replaces{
						{
							Filepath: `(\.go|\.proto|go.mod)$`,
							Old:      "placeholder",
							New:      "placereleaser",
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			template, err := Read(tc.path, tc.params)

			if tc.expectedError != "" {
				assert.Empty(t, template)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTemplate, template)
			}
		})
	}
}

func TestChanges_execute(t *testing.T) {
	tests := []struct {
		name          string
		changes       Changes
		root          string
		expectedError string
	}{
		{
			name: "Empty",
			changes: Changes{
				Deletes:  Deletes{},
				Moves:    Moves{},
				Appends:  Appends{},
				Replaces: Replaces{},
			},
			root:          "./test",
			expectedError: "",
		},
		{
			name: "MoveFails",
			changes: Changes{
				Deletes: Deletes{
					{Glob: "["},
				},
			},
			root:          "./test",
			expectedError: "syntax error in pattern",
		},
		{
			name: "MoveFails",
			changes: Changes{
				Deletes: Deletes{
					{Glob: "temp/foo"},
				},
				Moves: Moves{
					{
						Src: "temp/baz",
					},
				},
			},
			root:          "./test",
			expectedError: "rename test/temp/baz test: no such file or directory",
		},
		{
			name: "MoveFails",
			changes: Changes{
				Deletes: Deletes{
					{Glob: "temp/foo"},
				},
				Moves: Moves{
					{
						Src: "temp/baz",
					},
				},
			},
			root:          "./test",
			expectedError: "rename test/temp/baz test: no such file or directory",
		},
		{
			name: "AppendFails",
			changes: Changes{
				Deletes: Deletes{
					{Glob: "temp/foo"},
				},
				Moves: Moves{
					{
						Src:  "temp/bar",
						Dest: "temp/baz",
					},
				},
				Appends: Appends{
					{
						Filepath: "/",
					},
				},
			},
			root:          "./test",
			expectedError: "open test: is a directory",
		},
		{
			name: "ReplaceFails",
			changes: Changes{
				Deletes: Deletes{
					{Glob: "temp/foo"},
				},
				Moves: Moves{
					{
						Src:  "temp/bar",
						Dest: "temp/baz",
					},
				},
				Appends: Appends{
					{
						Filepath: "temp/baz",
						Content:  "More content",
					},
				},
				Replaces: Replaces{
					{
						Filepath: "[",
					},
				},
			},
			root:          "./test",
			expectedError: "error parsing regexp: missing closing ]: `[`",
		},
		{
			name: "Success",
			changes: Changes{
				Deletes: Deletes{
					{Glob: "temp/foo"},
				},
				Moves: Moves{
					{
						Src:  "temp/bar",
						Dest: "temp/baz",
					},
				},
				Appends: Appends{
					{
						Filepath: "temp/baz",
						Content:  "More content",
					},
				},
				Replaces: Replaces{
					{
						Filepath: "temp/ba*",
						Old:      "Lorem ipsum",
						New:      "Excepteur sint",
					},
				},
			},
			root:          "./test",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := prepare(t, tc.root)
			defer cleanup()

			u := ui.NewNop()
			err := tc.changes.execute(tc.root, u)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestDeletes_execute(t *testing.T) {
	tests := []struct {
		name          string
		deletes       Deletes
		root          string
		expectedError string
	}{
		{
			name:          "Empty",
			deletes:       Deletes{},
			root:          "./test",
			expectedError: "",
		},
		{
			name: "InvalidGlob",
			deletes: Deletes{
				{Glob: "["},
			},
			root:          "./test",
			expectedError: "syntax error in pattern",
		},
		{
			name: "Success",
			deletes: Deletes{
				{Glob: "temp/foo"},
				{Glob: "temp/ba*"},
			},
			root:          "./test",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := prepare(t, tc.root)
			defer cleanup()

			u := ui.NewNop()
			err := tc.deletes.execute(tc.root, u)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMoves_execute(t *testing.T) {
	tests := []struct {
		name          string
		moves         Moves
		root          string
		expectedError string
	}{
		{
			name:          "Empty",
			moves:         Moves{},
			root:          "./test",
			expectedError: "",
		},
		{
			name: "InvalidSource",
			moves: Moves{
				{
					Src: "temp/baz",
				},
			},
			root:          "./test",
			expectedError: "rename test/temp/baz test: no such file or directory",
		},
		{
			name: "Success",
			moves: Moves{
				{
					Src:  "temp/foo",
					Dest: "temp/bar",
				},
				{
					Src:  "temp/bar",
					Dest: "temp/baz",
				},
			},
			root:          "./test",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := prepare(t, tc.root)
			defer cleanup()

			u := ui.NewNop()
			err := tc.moves.execute(tc.root, u)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestAppends_execute(t *testing.T) {
	tests := []struct {
		name          string
		appends       Appends
		root          string
		expectedError string
	}{
		{
			name:          "Empty",
			appends:       Appends{},
			root:          "./test",
			expectedError: "",
		},
		{
			name: "InvalidFilepath",
			appends: Appends{
				{
					Filepath: "/",
				},
			},
			root:          "./test",
			expectedError: "open test: is a directory",
		},
		{
			name: "Success_Append",
			appends: Appends{
				{
					Filepath: "temp/foo",
					Content:  "More content",
				},
				{
					Filepath: "temp/bar",
					Content:  "More content",
				},
			},
			root:          "./test",
			expectedError: "",
		},
		{
			name: "Success_Create",
			appends: Appends{
				{
					Filepath: "temp/baz",
					Content:  "More content",
				},
			},
			root:          "./test",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := prepare(t, tc.root)
			defer cleanup()

			u := ui.NewNop()
			err := tc.appends.execute(tc.root, u)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestReplaces_execute(t *testing.T) {
	tests := []struct {
		name          string
		replaces      Replaces
		root          string
		expectedError string
	}{
		{
			name:          "Empty",
			replaces:      Replaces{},
			root:          "./test",
			expectedError: "",
		},
		{
			name: "InvalidFilepath",
			replaces: Replaces{
				{
					Filepath: "[",
				},
			},
			root:          "./test",
			expectedError: "error parsing regexp: missing closing ]: `[`",
		},
		{
			name: "Success",
			replaces: Replaces{
				{
					Filepath: "temp/foo",
					Old:      "Lorem ipsum",
					New:      "Excepteur sint",
				},
				{
					Filepath: "temp/ba*",
					Old:      "Lorem ipsum",
					New:      "Excepteur sint",
				},
			},
			root:          "./test",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := prepare(t, tc.root)
			defer cleanup()

			u := ui.NewNop()
			err := tc.replaces.execute(tc.root, u)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
