package shell_test

import (
	"context"
	"fmt"

	"github.com/gardenbed/basil-cli/internal/shell"
)

func ExampleRun() {
	_, out, _ := shell.Run(context.Background(), "echo", "foo", "bar")
	fmt.Println(out)
}

func ExampleRunWith() {
	opts := shell.RunOptions{
		Environment: map[string]string{
			"PLACEHOLDER": "foo bar",
		},
	}

	_, out, _ := shell.RunWith(context.Background(), opts, "printenv", "PLACEHOLDER")
	fmt.Println(out)
}

func ExampleRunner() {
	echo := shell.Runner("echo", "foo", "bar")
	_, out, _ := echo(context.Background(), "baz")
	fmt.Println(out)
}

func ExampleRunnerFunc_WithArgs() {
	echo := shell.Runner("echo", "foo")
	echo = echo.WithArgs("bar")
	_, out, _ := echo(context.Background(), "baz")
	fmt.Println(out)
}

func ExampleRunnerWith() {
	opts := shell.RunOptions{
		Environment: map[string]string{
			"TOKEN": "access-token",
		},
	}

	printenv := shell.RunnerWith("printenv")
	_, out, _ := printenv(context.Background(), opts, "TOKEN")
	fmt.Println(out)
}

func ExampleRunnerWithFunc_WithArgs() {
	opts := shell.RunOptions{
		Environment: map[string]string{
			"TOKEN": "access-token",
		},
	}

	printenv := shell.RunnerWith("printenv")
	printenv = printenv.WithArgs("TOKEN")
	_, out, _ := printenv(context.Background(), opts)
	fmt.Println(out)
}
