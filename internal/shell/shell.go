// Package shell provides helper functions for running shell commands.
package shell

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunOptions are optional settings for a command.
type RunOptions struct {
	// WorkingDir is the working directory for a command.
	WorkingDir string
	// Environment is a map of key-values representing environment variables for a command.
	Environment map[string]string
}

func run(ctx context.Context, opts RunOptions, command string, args ...string) (int, string, error) {
	var cmdExitCode int
	var cmdOutput string
	var cmdError error

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = opts.WorkingDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = os.Environ()
	for key, val := range opts.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}

	err := cmd.Run()

	if err != nil {
		cmdError = fmt.Errorf("error on running %s: %s: %s",
			strings.Join(append([]string{command}, args...), " "),
			err,
			strings.Trim(stderr.String(), "\n"),
		)

		if exitErr, ok := err.(*exec.ExitError); ok {
			cmdExitCode = exitErr.ExitCode()
		} else {
			cmdExitCode = -1 // Unknown exit code
		}
	} else {
		cmdOutput = strings.Trim(stdout.String(), "\n")
	}

	return cmdExitCode, cmdOutput, cmdError
}

// Run executes a command in the default shell.
// It returns the exit code, output, and error (if any).
func Run(ctx context.Context, command string, args ...string) (int, string, error) {
	return run(ctx, RunOptions{}, command, args...)
}

// RunWith executes a command with given options in the default shell.
// It returns the exit code, output, and error (if any).
func RunWith(ctx context.Context, opts RunOptions, command string, args ...string) (int, string, error) {
	return run(ctx, opts, command, args...)
}

// RunnerFunc is a function for running a bounded command.
type RunnerFunc func(context.Context, ...string) (int, string, error)

// Runner binds a command to a list of arguments and returns a function.
// The returned function can be used for running the bounded command in the default shell.
func Runner(command string, args ...string) RunnerFunc {
	return func(ctx context.Context, a ...string) (int, string, error) {
		all := append(args, a...)
		return run(ctx, RunOptions{}, command, all...)
	}
}

// WithArgs binds more arguments to a Runner function and returns a new Runner function.
func (f RunnerFunc) WithArgs(args ...string) RunnerFunc {
	return func(ctx context.Context, a ...string) (int, string, error) {
		all := append(args, a...)
		return f(ctx, all...)
	}
}

// RunnerWithFunc is a function for running a bounded command with given options.
type RunnerWithFunc func(context.Context, RunOptions, ...string) (int, string, error)

// RunnerWith binds a command to a list of arguments with given options and returns a function.
// The returned function can be used for running the bounded command in the default shell.
func RunnerWith(command string, args ...string) RunnerWithFunc {
	return func(ctx context.Context, opts RunOptions, a ...string) (int, string, error) {
		all := append(args, a...)
		return run(ctx, opts, command, all...)
	}
}

// WithArgs binds more arguments to a RunnerWith function and returns a new RunnerWith function.
func (f RunnerWithFunc) WithArgs(args ...string) RunnerWithFunc {
	return func(ctx context.Context, opts RunOptions, a ...string) (int, string, error) {
		all := append(args, a...)
		return f(ctx, opts, all...)
	}
}
