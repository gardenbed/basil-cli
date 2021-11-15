package compile

import "github.com/gardenbed/basil-cli/internal/ui"

// Compiler is used for parsing Go source code files and compiling new source code files.
type Compiler struct {
	parser *parser
}

// New creates a new compiler.
// This is meant to be used by downstream packages that provide Consumer.
func New(ui ui.UI, consumers ...*Consumer) *Compiler {
	return &Compiler{
		parser: &parser{
			ui:        ui,
			consumers: consumers,
		},
	}
}

// Compile parses all Go source code files in the given packages and generates new artifacts (source codes).
func (c *Compiler) Compile(packages string, opts ParseOptions) error {
	return c.parser.Parse(packages, opts)
}
