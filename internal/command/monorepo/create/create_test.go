package create

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/gardenbed/go-github"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
)

func TestNew(t *testing.T) {
	ui := cli.NewMockUi()
	config := config.Config{}
	c := New(ui, config)

	assert.NotNil(t, c)
}

func TestNewFactory(t *testing.T) {
	ui := cli.NewMockUi()
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
		c := &Command{ui: cli.NewMockUi()}
		exitCode := c.Run([]string{"-undefined"})

		assert.Equal(t, command.FlagError, exitCode)
	})

	t.Run("OK", func(t *testing.T) {
		mockUI := cli.NewMockUi()
		mockUI.InputReader = bufio.NewReader(strings.NewReader(""))
		c := &Command{ui: mockUI}
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
			c := &Command{ui: cli.NewMockUi()}
			exitCode := c.parseFlags(tc.args)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_exec(t *testing.T) {
	tests := []struct {
		name             string
		repo             *MockRepoService
		archive          *MockArchiveService
		template         *MockTemplateService
		inputs           string
		expectedExitCode int
	}{
		{
			name:             "InvalidName",
			inputs:           "",
			expectedExitCode: command.InputError,
		},
		{
			name:             "EmptyName",
			inputs:           "\n",
			expectedExitCode: command.InputError,
		},
		{
			name:             "InvalidName",
			inputs:           "test monorepo\n",
			expectedExitCode: command.InputError,
		},
		{
			name: "DownloadFails",
			repo: &MockRepoService{
				DownloadTarArchiveMocks: []DownloadTarArchiveMock{
					{OutError: errors.New("github error")},
				},
			},
			inputs:           "test-monorepo\n",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "ExtractFails",
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
			inputs:           "test-monorepo\n",
			expectedExitCode: command.ArchiveError,
		},
		{
			name: "TemplateReadFails",
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
			inputs:           "test\n",
			expectedExitCode: command.TemplateError,
		},
		{
			name: "TemplateExecuteFails",
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
			inputs:           "test-monorepo\n",
			expectedExitCode: command.TemplateError,
		},
		{
			name: "Success",
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
			inputs:           "test-monorepo\n",
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// The Ask method creates a new bufio.Reader every time.
			// Simply assigning an strings.Reader to mockUI.InputReader causes the bufio.Reader.ReadString() to error the second time the Ask method is called.
			// We need to assign a bufio.Reader to mockUI.InputReader, so bufio.NewReader(), called in Ask, will reuse it instead of creating a new one.
			var inputReader io.Reader
			inputReader = strings.NewReader(tc.inputs)
			inputReader = bufio.NewReader(inputReader)

			mockUI := cli.NewMockUi()
			mockUI.InputReader = inputReader

			c := &Command{
				ui: mockUI,
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
