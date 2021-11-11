package compile

import "github.com/gardenbed/basil-cli/internal/debug"

// Compiler is used for parsing Go source code files and compiling new source code files.
type Compiler struct {
	parser *parser
}

// New creates a new compiler.
// This is meant to be used by downstream packages that provide Consumer.
func New(debugger *debug.DebuggerSet, consumers ...*Consumer) *Compiler {
	return &Compiler{
		parser: &parser{
			debugger:  debugger,
			consumers: consumers,
		},
	}
}

// Compile parses all Go source code files in the given packages and generates new artifacts (source codes).
func (c *Compiler) Compile(packages string, opts ParseOptions) error {
	return c.parser.Parse(packages, opts)
}
