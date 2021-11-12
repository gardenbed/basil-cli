package build

import "github.com/gardenbed/basil-cli/internal/compile"

type (
	CompileMock struct {
		InPackages string
		InOptions  compile.ParseOptions
		OutError   error
	}

	MockCompilerService struct {
		CompileIndex int
		CompileMocks []CompileMock
	}
)

func (m *MockCompilerService) Compile(packages string, opts compile.ParseOptions) error {
	i := m.CompileIndex
	m.CompileIndex++
	m.CompileMocks[i].InPackages = packages
	m.CompileMocks[i].InOptions = opts
	return m.CompileMocks[i].OutError
}
