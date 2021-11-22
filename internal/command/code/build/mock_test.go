package build

import "github.com/gardenbed/go-parser"

type (
	CompileMock struct {
		InPackages string
		InOptions  parser.ParseOptions
		OutError   error
	}

	MockCompilerService struct {
		CompileIndex int
		CompileMocks []CompileMock
	}
)

func (m *MockCompilerService) Compile(packages string, opts parser.ParseOptions) error {
	i := m.CompileIndex
	m.CompileIndex++
	m.CompileMocks[i].InPackages = packages
	m.CompileMocks[i].InOptions = opts
	return m.CompileMocks[i].OutError
}
