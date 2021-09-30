package command

const (
	// Success is the exit code when a command execution is successful.
	Success int = iota
	// SpecError is the exit code when reading the spec file fails.
	SpecError
	// FlagError is the exit code when an undefined or invalid flag is provided to a command.
	FlagError
	// OSError is the exit code when an OS operation fails.
	OSError
	// GoError is the exit code when a go command fails.
	GoError
	// GitError is the exit code when a git command fails.
	GitError
	// GitHubError is the exit code when a GitHub operation fails.
	GitHubError
	// ChangelogError is the exit code when generating the changelog fails.
	ChangelogError
)
