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

func TestParams_Has(t *testing.T) {
	tests := []struct {
		name           string
		params         Params
		param          string
		expectedResult bool
	}{
		{
			name:           "Found",
			params:         Params{"Name", "Owner"},
			param:          "Owner",
			expectedResult: true,
		},
		{
			name:           "NotFound",
			params:         Params{"Name", "Owner"},
			param:          "DockerID",
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.params.Has(tc.param)

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestTemplate_Execute(t *testing.T) {
	tests := []struct {
		name          string
		template      Template
		root          string
		expectedError string
	}{
		{
			name: "Empty",
			template: Template{
				Edits: Edits{
					Deletes:  Deletes{},
					Moves:    Moves{},
					Appends:  Appends{},
					Replaces: Replaces{},
				},
			},
			root:          "./test",
			expectedError: "",
		},
		{
			name: "MoveFails",
			template: Template{
				Edits: Edits{
					Deletes: Deletes{
						{Glob: "["},
					},
				},
			},
			root:          "./test",
			expectedError: "syntax error in pattern",
		},
		{
			name: "MoveFails",
			template: Template{
				Edits: Edits{
					Deletes: Deletes{
						{Glob: "temp/foo"},
					},
					Moves: Moves{
						{
							Src: "temp/baz",
						},
					},
				},
			},
			root:          "./test",
			expectedError: "rename test/temp/baz test: no such file or directory",
		},
		{
			name: "MoveFails",
			template: Template{
				Edits: Edits{
					Deletes: Deletes{
						{Glob: "temp/foo"},
					},
					Moves: Moves{
						{
							Src: "temp/baz",
						},
					},
				},
			},
			root:          "./test",
			expectedError: "rename test/temp/baz test: no such file or directory",
		},
		{
			name: "AppendFails",
			template: Template{
				Edits: Edits{
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
			},
			root:          "./test",
			expectedError: "open test: is a directory",
		},
		{
			name: "ReplaceFails",
			template: Template{
				Edits: Edits{
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
			},
			root:          "./test",
			expectedError: "error parsing regexp: missing closing ]: `[`",
		},
		{
			name: "Success",
			template: Template{
				Edits: Edits{
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
			err := tc.template.Execute(u, tc.root)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestEdits_execute(t *testing.T) {
	tests := []struct {
		name          string
		edits         Edits
		root          string
		expectedError string
	}{
		{
			name: "Empty",
			edits: Edits{
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
			edits: Edits{
				Deletes: Deletes{
					{Glob: "["},
				},
			},
			root:          "./test",
			expectedError: "syntax error in pattern",
		},
		{
			name: "MoveFails",
			edits: Edits{
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
			edits: Edits{
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
			edits: Edits{
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
			edits: Edits{
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
			edits: Edits{
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
			err := tc.edits.execute(u, tc.root)

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
			err := tc.deletes.execute(u, tc.root)

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
			err := tc.moves.execute(u, tc.root)

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
			err := tc.appends.execute(u, tc.root)

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
			err := tc.replaces.execute(u, tc.root)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
