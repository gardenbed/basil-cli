package create

import (
	"errors"
	"testing"

	"github.com/gardenbed/go-github"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/template"
	"github.com/gardenbed/basil-cli/internal/ui"
)

func TestNew(t *testing.T) {
	ui := ui.NewNop()
	config := config.Config{}
	c := New(ui, config)

	assert.NotNil(t, c)
}

func TestNewFactory(t *testing.T) {
	ui := ui.NewNop()
	config := config.Config{}
	c, err := NewFactory(ui, config)()

	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestCommand_Synopsis(t *testing.T) {
	c := new(Command)
	synopsis := c.Synopsis()

	assert.NotEmpty(t, synopsis)
}

func TestCommand_Help(t *testing.T) {
	c := new(Command)
	help := c.Help()

	assert.NotEmpty(t, help)
}

func TestCommand_Run(t *testing.T) {
	t.Run("InvalidFlag", func(t *testing.T) {
		c := &Command{ui: ui.NewNop()}
		exitCode := c.Run([]string{"-undefined"})

		assert.Equal(t, command.FlagError, exitCode)
	})

	t.Run("OK", func(t *testing.T) {
		c := &Command{ui: &MockUI{
			UI: ui.NewNop(),
			AskMocks: []AskMock{
				{OutError: errors.New("io error")},
			},
		}}

		c.Run([]string{})

		assert.NotNil(t, c.services.repo)
		assert.NotNil(t, c.services.archive)
		assert.NotNil(t, c.services.template)
	})
}

func TestCommand_parseFlags(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedExitCode int
	}{
		{
			name:             "InvalidFlag",
			args:             []string{"-undefined"},
			expectedExitCode: command.FlagError,
		},
		{
			name:             "NoFlag",
			args:             []string{},
			expectedExitCode: command.Success,
		},
		{
			name: "ValidFlags",
			args: []string{
				"-name", "test-monorepo",
				"-revision", "test",
			},
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{ui: ui.NewNop()}
			exitCode := c.parseFlags(tc.args)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_exec(t *testing.T) {
	tests := []struct {
		name             string
		ui               *MockUI
		repo             *MockRepoService
		archive          *MockArchiveService
		template         *MockTemplateService
		expectedExitCode int
	}{
		{
			name: "AskFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutError: errors.New("io error")},
				},
			},
			expectedExitCode: command.InputError,
		},
		{
			name: "DownloadFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-monorepo"},
				},
			},
			repo: &MockRepoService{
				DownloadTarArchiveMocks: []DownloadTarArchiveMock{
					{OutError: errors.New("github error")},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name: "ExtractFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-monorepo"},
				},
			},
			repo: &MockRepoService{
				DownloadTarArchiveMocks: []DownloadTarArchiveMock{
					{OutResponse: &github.Response{}},
				},
			},
			archive: &MockArchiveService{
				ExtractMocks: []ExtractMock{
					{OutError: errors.New("archive error")},
				},
			},
			expectedExitCode: command.ArchiveError,
		},
		{
			name: "TemplateLoadFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-monorepo"},
				},
			},
			repo: &MockRepoService{
				DownloadTarArchiveMocks: []DownloadTarArchiveMock{
					{OutResponse: &github.Response{}},
				},
			},
			archive: &MockArchiveService{
				ExtractMocks: []ExtractMock{
					{OutError: nil},
				},
			},
			template: &MockTemplateService{
				LoadMocks: []LoadMock{
					{OutError: errors.New("template error")},
				},
			},
			expectedExitCode: command.TemplateError,
		},
		{
			name: "TemplateFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-monorepo"},
				},
			},
			repo: &MockRepoService{
				DownloadTarArchiveMocks: []DownloadTarArchiveMock{
					{OutResponse: &github.Response{}},
				},
			},
			archive: &MockArchiveService{
				ExtractMocks: []ExtractMock{
					{OutError: nil},
				},
			},
			template: &MockTemplateService{
				LoadMocks: []LoadMock{
					{OutError: nil},
				},
				ParamsMocks: []ParamsMock{
					{OutParams: template.Params{"Name"}},
				},
				TemplateMocks: []TemplateMock{
					{OutError: errors.New("template error")},
				},
			},
			expectedExitCode: command.TemplateError,
		},
		{
			name: "TemplateExecuteFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-monorepo"},
				},
			},
			repo: &MockRepoService{
				DownloadTarArchiveMocks: []DownloadTarArchiveMock{
					{OutResponse: &github.Response{}},
				},
			},
			archive: &MockArchiveService{
				ExtractMocks: []ExtractMock{
					{OutError: nil},
				},
			},
			template: &MockTemplateService{
				LoadMocks: []LoadMock{
					{OutError: nil},
				},
				ParamsMocks: []ParamsMock{
					{OutParams: template.Params{"Name"}},
				},
				TemplateMocks: []TemplateMock{
					{
						OutTemplate: &template.Template{
							Edits: template.Edits{
								Deletes: template.Deletes{
									{Glob: "["},
								},
							},
						},
					},
				},
			},
			expectedExitCode: command.TemplateError,
		},
		{
			name: "Success",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-monorepo"},
				},
			},
			repo: &MockRepoService{
				DownloadTarArchiveMocks: []DownloadTarArchiveMock{
					{OutResponse: &github.Response{}},
				},
			},
			archive: &MockArchiveService{
				ExtractMocks: []ExtractMock{
					{OutError: nil},
				},
			},
			template: &MockTemplateService{
				LoadMocks: []LoadMock{
					{OutError: nil},
				},
				ParamsMocks: []ParamsMock{
					{OutParams: template.Params{"Name"}},
				},
				TemplateMocks: []TemplateMock{
					{
						OutTemplate: &template.Template{
							Edits: template.Edits{
								Deletes:  template.Deletes{},
								Moves:    template.Moves{},
								Appends:  template.Appends{},
								Replaces: template.Replaces{},
							},
						},
					},
				},
			},
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui: tc.ui,
			}

			c.services.repo = tc.repo
			c.services.archive = tc.archive
			c.services.template = tc.template

			exitCode := c.exec()

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestSelectTemplatePath(t *testing.T) {
	tests := []struct {
		name         string
		nameFlag     string
		path         string
		expectedPath string
		expectedBool bool
	}{
		{
			name:         "NonTemplatePath",
			nameFlag:     "go-monorepo",
			path:         "gardenbed-basil-templates-0abcdef/file",
			expectedPath: "",
			expectedBool: false,
		},
		{
			name:         "TemplatePath",
			nameFlag:     "go-monorepo",
			path:         "gardenbed-basil-templates-0abcdef/go/monorepo/file",
			expectedPath: "go-monorepo/file",
			expectedBool: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{}
			c.flags.name = tc.nameFlag

			path, b := c.selectTemplatePath(tc.path)

			assert.Equal(t, tc.expectedPath, path)
			assert.Equal(t, tc.expectedBool, b)
		})
	}
}

func TestValidateInputName(t *testing.T) {
	tests := []struct {
		name          string
		val           string
		expectedError string
	}{
		{
			name:          "InvalidName",
			val:           "test monorepo",
			expectedError: "invalid name: test monorepo",
		},
		{
			name:          "ValidName",
			val:           "test-monorepo",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInputName(tc.val)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
