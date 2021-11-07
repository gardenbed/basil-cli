package create

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/go-github"
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
				"-name", "my-service",
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
			inputs:           "my service\n",
			expectedExitCode: command.InputError,
		},
		{
			name:             "InvalidOwner",
			inputs:           "my-service\n",
			expectedExitCode: command.InputError,
		},
		{
			name:             "EmptyOwner",
			inputs:           "my-service\n\n",
			expectedExitCode: command.InputError,
		},
		{
			name:             "InvalidOwner",
			inputs:           "my-service\nmy team\n",
			expectedExitCode: command.InputError,
		},
		{
			name:             "InvalidProfile",
			inputs:           "my-service\nmy-team\n",
			expectedExitCode: command.InputError,
		},
		{
			name:             "EmptyProfile",
			inputs:           "my-service\nmy-team\n\n",
			expectedExitCode: command.InputError,
		},
		{
			name:             "InvalidProfile",
			inputs:           "my-service\nmy-team\nunknown-profile\n",
			expectedExitCode: command.InputError,
		},
		{
			name:             "InvalidDockerID",
			inputs:           "my-service\nmy-team\ngrpc-service\n",
			expectedExitCode: command.InputError,
		},
		{
			name:             "EmptyDockerID",
			inputs:           "my-service\nmy-team\ngrpc-service\n\n",
			expectedExitCode: command.InputError,
		},
		{
			name:             "InvalidDockerID",
			inputs:           "my-service\nmy-team\ngrpc-service\ndocker id\n",
			expectedExitCode: command.InputError,
		},
		{
			name: "DownloadFails",
			repo: &MockRepoService{
				DownloadTarArchiveMocks: []DownloadTarArchiveMock{
					{OutError: errors.New("github error")},
				},
			},
			inputs:           "my-service\nmy-team\ngrpc-service\norca\n",
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
			inputs:           "my-service\nmy-team\ngrpc-service\norca\n",
			expectedExitCode: command.ArchiveError,
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
			inputs:           "my-service\nmy-team\ngrpc-service\norca\n",
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

			exitCode := c.exec()

			assert.Equal(t, tc.expectedExitCode, exitCode)
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
			nameFlag:     "my-service",
			profileFlag:  "grpc-service",
			path:         "gardenbed-basil-templates-0abcdef/file",
			expectedPath: "",
			expectedBool: false,
		},
		{
			name:         "TemplatePath",
			nameFlag:     "my-service",
			profileFlag:  "grpc-service",
			path:         "gardenbed-basil-templates-0abcdef/go/grpc-service/file",
			expectedPath: "my-service/file",
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
