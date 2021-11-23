package template

import (
	"testing"

	"github.com/gardenbed/charm/ui"
	"github.com/stretchr/testify/assert"
)

const invalidYAMLTemplate = `
name: test-template
description: This template is used for testing.

edits:
  replaces:
    - filepath: 'README.md$'
      old: 'placeholder'
      new: '{{'
`

const unknownTemplateParam = `
name: test-template
description: This template is used for testing.

edits:
  replaces:
    - filepath: 'README.md$'
      old: 'placeholder'
      new: '{{.Unknown}}'
`

const validYAMLTemplate = `
name: test-template
description: This template is used for testing.

edits:
  deletes:
    - glob: 'template.yaml'
  moves:
    - src: './cmd/placeholder'
      dest: './cmd/{{.Name}}'
  appends:
    - filepath: './.github/CODEOWNERS'
      content: '@octocat'
  replaces:
    - filepath: '(\.go|\.proto|go.mod)$'
      old: 'placeholder'
      new: '{{.Name}}'
`

func TestNewService(t *testing.T) {
	ui := ui.NewNop()
	s := NewService(ui)

	assert.NotNil(t, s)
	assert.NotNil(t, s.ui)
}

func TestService_Load(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectedError string
	}{
		{
			name:          "NoFile",
			path:          "./test",
			expectedError: "template file not found",
		},
		{
			name:          "Success",
			path:          "./test/valid",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{
				ui: ui.NewNop(),
			}

			err := s.Load(tc.path)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestService_Params(t *testing.T) {
	tests := []struct {
		name   string
		params Params
	}{
		{
			name:   "OK",
			params: Params{"Name", "Owner"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{
				ui:     ui.NewNop(),
				params: tc.params,
			}

			params := s.Params()

			assert.Equal(t, tc.params, params)
		})
	}
}

func TestService_Template(t *testing.T) {
	tests := []struct {
		name             string
		text             string
		inputs           interface{}
		expectedTemplate *Template
		expectedError    string
	}{
		{
			name:          "EmptyYAML",
			text:          ``,
			inputs:        nil,
			expectedError: "EOF",
		},
		{
			name:          "InvalidYAML",
			text:          `invalid yaml`,
			inputs:        nil,
			expectedError: "yaml: unmarshal errors",
		},
		{
			name:          "InvalidTemplate",
			text:          invalidYAMLTemplate,
			inputs:        nil,
			expectedError: "unterminated character constant",
		},
		{
			name:          "UnknownParam",
			text:          unknownTemplateParam,
			inputs:        struct{}{},
			expectedError: "can't evaluate field Unknown in type struct",
		},
		{
			name: "Success",
			text: validYAMLTemplate,
			inputs: struct {
				Name string
			}{
				Name: "placereleaser",
			},
			expectedTemplate: &Template{
				Name:        "test-template",
				Description: "This template is used for testing.",
				Edits: Edits{
					Deletes: Deletes{
						{Glob: "template.yaml"},
					},
					Moves: Moves{
						{
							Src:  "./cmd/placeholder",
							Dest: "./cmd/placereleaser",
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
			s := &Service{
				ui:   ui.NewNop(),
				text: tc.text,
			}

			template, err := s.Template(tc.inputs)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTemplate, template)
			} else {
				assert.Nil(t, template)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}
