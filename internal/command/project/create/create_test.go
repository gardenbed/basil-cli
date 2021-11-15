package create

import (
	"errors"
	"testing"

	"github.com/gardenbed/go-github"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
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
				"-name", "test-project",
				"-owner", "my-team",
				"-profile", "grpc-service",
				"-dockerid", "orca",
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
			name: "AskNameFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutError: errors.New("io error")},
				},
			},
			expectedExitCode: command.InputError,
		},
		{
			name: "AskOwnerFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-project"},
					{OutError: errors.New("io error")},
				},
			},
			expectedExitCode: command.InputError,
		},
		{
			name: "SelectProfileFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-project"},
					{OutValue: "my-team"},
				},
				SelectMocks: []SelectMock{
					{OutError: errors.New("io error")},
				},
			},
			expectedExitCode: command.InputError,
		},
		{
			name: "AskDockerIDFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-project"},
					{OutValue: "my-team"},
					{OutError: errors.New("io error")},
				},
				SelectMocks: []SelectMock{
					{
						OutItem: ui.Item{
							Key: "grpc-service",
						},
					},
				},
			},
			expectedExitCode: command.InputError,
		},
		{
			name: "DownloadFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-project"},
					{OutValue: "my-team"},
					{OutValue: "orca"},
				},
				SelectMocks: []SelectMock{
					{
						OutItem: ui.Item{
							Key: "grpc-service",
						},
					},
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
					{OutValue: "test-project"},
					{OutValue: "my-team"},
					{OutValue: "orca"},
				},
				SelectMocks: []SelectMock{
					{
						OutItem: ui.Item{
							Key: "grpc-service",
						},
					},
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
			name: "TemplateReadFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "invalid"},
					{OutValue: "my-team"},
					{OutValue: "orca"},
				},
				SelectMocks: []SelectMock{
					{
						OutItem: ui.Item{
							Key: "grpc-service",
						},
					},
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
			expectedExitCode: command.TemplateError,
		},
		{
			name: "TemplateExecuteFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-project"},
					{OutValue: "my-team"},
					{OutValue: "orca"},
				},
				SelectMocks: []SelectMock{
					{
						OutItem: ui.Item{
							Key: "grpc-service",
						},
					},
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
				ExecuteMocks: []ExecuteMock{
					{OutError: errors.New("template error")},
				},
			},
			expectedExitCode: command.TemplateError,
		},
		{
			name: "Success",
			ui: &MockUI{
				UI: ui.NewNop(),
				AskMocks: []AskMock{
					{OutValue: "test-project"},
					{OutValue: "my-team"},
					{OutValue: "orca"},
				},
				SelectMocks: []SelectMock{
					{
						OutItem: ui.Item{
							Key: "grpc-service",
						},
					},
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
				ExecuteMocks: []ExecuteMock{
					{OutError: nil},
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

func TestValidateInputName(t *testing.T) {
	tests := []struct {
		name          string
		val           string
		expectedError string
	}{
		{
			name:          "InvalidName",
			val:           "test project",
			expectedError: "invalid name: test project",
		},
		{
			name:          "ValidName",
			val:           "test-project",
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

func TestValidateInputOwner(t *testing.T) {
	tests := []struct {
		name          string
		val           string
		expectedError string
	}{
		{
			name:          "InvalidOwner",
			val:           "my team",
			expectedError: "invalid owner: my team",
		},
		{
			name:          "ValidOwner",
			val:           "my-team",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInputOwner(tc.val)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestSearchProfile(t *testing.T) {
	tests := []struct {
		name           string
		val            string
		index          int
		expectedResult bool
	}{
		{
			name:           "Found",
			val:            "grpc",
			index:          2,
			expectedResult: true,
		},
		{
			name:           "NotFound",
			val:            "thrift",
			index:          2,
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := searchProfile(tc.val, tc.index)

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestValidateInputDockerID(t *testing.T) {
	tests := []struct {
		name          string
		val           string
		expectedError string
	}{
		{
			name:          "InvalidDockerID",
			val:           "docker id",
			expectedError: "invalid Docker ID: docker id",
		},
		{
			name:          "ValidDockerID",
			val:           "orca",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInputDockerID(tc.val)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestSelectTemplatePath(t *testing.T) {
	tests := []struct {
		name         string
		nameFlag     string
		profileFlag  string
		path         string
		expectedPath string
		expectedBool bool
	}{
		{
			name:         "NonTemplatePath",
			nameFlag:     "test-project",
			profileFlag:  "grpc-service",
			path:         "gardenbed-basil-templates-0abcdef/file",
			expectedPath: "",
			expectedBool: false,
		},
		{
			name:         "TemplatePath",
			nameFlag:     "test-project",
			profileFlag:  "grpc-service",
			path:         "gardenbed-basil-templates-0abcdef/go/grpc-service/file",
			expectedPath: "test-project/file",
			expectedBool: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{}
			c.flags.name = tc.nameFlag
			c.flags.profile = tc.profileFlag

			path, b := c.selectTemplatePath()(tc.path)

			assert.Equal(t, tc.expectedPath, path)
			assert.Equal(t, tc.expectedBool, b)
		})
	}
}
