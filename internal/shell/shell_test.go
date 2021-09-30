package shell

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name             string
		ctx              context.Context
		command          string
		args             []string
		expectedExitCode int
		expectedOutput   string
		expectedError    error
	}{
		{
			name:             "NotFound",
			ctx:              context.Background(),
			command:          "unknown",
			args:             []string{},
			expectedExitCode: -1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running unknown: exec: \"unknown\": executable file not found in $PATH: "),
		},
		{
			name:             "Error",
			ctx:              context.Background(),
			command:          "cat",
			args:             []string{"null"},
			expectedExitCode: 1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running cat null: exit status 1: cat: null: No such file or directory"),
		},
		{
			name:             "Success",
			ctx:              context.Background(),
			command:          "echo",
			args:             []string{"foo", "bar"},
			expectedExitCode: 0,
			expectedOutput:   "foo bar",
			expectedError:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, out, err := Run(tc.ctx, tc.command, tc.args...)

			assert.Equal(t, tc.expectedExitCode, code)
			assert.Equal(t, tc.expectedOutput, out)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestRunWith(t *testing.T) {
	tests := []struct {
		name             string
		ctx              context.Context
		opts             RunOptions
		command          string
		args             []string
		expectedExitCode int
		expectedOutput   string
		expectedError    error
	}{
		{
			name:             "NotFound",
			ctx:              context.Background(),
			opts:             RunOptions{},
			command:          "unknown",
			args:             []string{},
			expectedExitCode: -1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running unknown: exec: \"unknown\": executable file not found in $PATH: "),
		},
		{
			name:             "Error",
			ctx:              context.Background(),
			opts:             RunOptions{},
			command:          "cat",
			args:             []string{"null"},
			expectedExitCode: 1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running cat null: exit status 1: cat: null: No such file or directory"),
		},
		{
			name: "Success",
			ctx:  context.Background(),
			opts: RunOptions{
				Environment: map[string]string{
					"TOKEN": "access-token",
				},
			},
			command:          "printenv",
			args:             []string{"TOKEN"},
			expectedExitCode: 0,
			expectedOutput:   "access-token",
			expectedError:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, out, err := RunWith(tc.ctx, tc.opts, tc.command, tc.args...)

			assert.Equal(t, tc.expectedExitCode, code)
			assert.Equal(t, tc.expectedOutput, out)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestRunner(t *testing.T) {
	tests := []struct {
		name             string
		command          string
		args             []string
		runCtx           context.Context
		runArgs          []string
		expectedExitCode int
		expectedOutput   string
		expectedError    error
	}{
		{
			name:             "NotFound",
			command:          "unknown",
			args:             []string{},
			runCtx:           context.Background(),
			runArgs:          []string{},
			expectedExitCode: -1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running unknown: exec: \"unknown\": executable file not found in $PATH: "),
		},
		{
			name:             "Error",
			command:          "cat",
			args:             []string{"null"},
			runCtx:           context.Background(),
			runArgs:          []string{},
			expectedExitCode: 1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running cat null: exit status 1: cat: null: No such file or directory"),
		},
		{
			name:             "Success",
			command:          "echo",
			args:             []string{"foo", "bar"},
			runCtx:           context.Background(),
			runArgs:          []string{"baz"},
			expectedExitCode: 0,
			expectedOutput:   "foo bar baz",
			expectedError:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			run := Runner(tc.command, tc.args...)
			code, out, err := run(tc.runCtx, tc.runArgs...)

			assert.Equal(t, tc.expectedExitCode, code)
			assert.Equal(t, tc.expectedOutput, out)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestRunnerFunc_WithArgs(t *testing.T) {
	tests := []struct {
		name             string
		f                RunnerFunc
		args             []string
		runCtx           context.Context
		runArgs          []string
		expectedExitCode int
		expectedOutput   string
		expectedError    error
	}{
		{
			name:             "NotFound",
			f:                Runner("unknown"),
			args:             []string{},
			runCtx:           context.Background(),
			runArgs:          []string{},
			expectedExitCode: -1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running unknown: exec: \"unknown\": executable file not found in $PATH: "),
		},
		{
			name:             "Error",
			f:                Runner("cat"),
			args:             []string{"null"},
			runCtx:           context.Background(),
			runArgs:          []string{},
			expectedExitCode: 1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running cat null: exit status 1: cat: null: No such file or directory"),
		},
		{
			name:             "Success",
			f:                Runner("echo"),
			args:             []string{"foo", "bar"},
			runCtx:           context.Background(),
			runArgs:          []string{"baz"},
			expectedExitCode: 0,
			expectedOutput:   "foo bar baz",
			expectedError:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			run := tc.f.WithArgs(tc.args...)
			code, out, err := run(tc.runCtx, tc.runArgs...)

			assert.Equal(t, tc.expectedExitCode, code)
			assert.Equal(t, tc.expectedOutput, out)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestRunnerWith(t *testing.T) {
	tests := []struct {
		name             string
		command          string
		args             []string
		runCtx           context.Context
		runOpts          RunOptions
		runArgs          []string
		expectedExitCode int
		expectedOutput   string
		expectedError    error
	}{
		{
			name:             "NotFound",
			command:          "unknown",
			args:             []string{},
			runCtx:           context.Background(),
			runOpts:          RunOptions{},
			runArgs:          []string{},
			expectedExitCode: -1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running unknown: exec: \"unknown\": executable file not found in $PATH: "),
		},
		{
			name:             "Error",
			command:          "cat",
			args:             []string{"null"},
			runCtx:           context.Background(),
			runOpts:          RunOptions{},
			runArgs:          []string{},
			expectedExitCode: 1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running cat null: exit status 1: cat: null: No such file or directory"),
		},
		{
			name:    "Success",
			command: "printenv",
			args:    []string{},
			runCtx:  context.Background(),
			runOpts: RunOptions{
				Environment: map[string]string{
					"TOKEN": "access-token",
				},
			},
			runArgs:          []string{"TOKEN"},
			expectedExitCode: 0,
			expectedOutput:   "access-token",
			expectedError:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			run := RunnerWith(tc.command, tc.args...)
			code, out, err := run(tc.runCtx, tc.runOpts, tc.runArgs...)

			assert.Equal(t, tc.expectedExitCode, code)
			assert.Equal(t, tc.expectedOutput, out)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestRunnerWithFunc_WithArgs(t *testing.T) {
	tests := []struct {
		name             string
		f                RunnerWithFunc
		args             []string
		runCtx           context.Context
		runOpts          RunOptions
		runArgs          []string
		expectedExitCode int
		expectedOutput   string
		expectedError    error
	}{
		{
			name:             "NotFound",
			f:                RunnerWith("unknown"),
			args:             []string{},
			runCtx:           context.Background(),
			runOpts:          RunOptions{},
			runArgs:          []string{},
			expectedExitCode: -1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running unknown: exec: \"unknown\": executable file not found in $PATH: "),
		},
		{
			name:             "Error",
			f:                RunnerWith("cat"),
			args:             []string{"null"},
			runCtx:           context.Background(),
			runOpts:          RunOptions{},
			runArgs:          []string{},
			expectedExitCode: 1,
			expectedOutput:   "",
			expectedError:    errors.New("error on running cat null: exit status 1: cat: null: No such file or directory"),
		},
		{
			name:             "Success",
			f:                RunnerWith("echo"),
			args:             []string{"foo", "bar"},
			runCtx:           context.Background(),
			runOpts:          RunOptions{},
			runArgs:          []string{"baz"},
			expectedExitCode: 0,
			expectedOutput:   "foo bar baz",
			expectedError:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			run := tc.f.WithArgs(tc.args...)
			code, out, err := run(tc.runCtx, tc.runOpts, tc.runArgs...)

			assert.Equal(t, tc.expectedExitCode, code)
			assert.Equal(t, tc.expectedOutput, out)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
