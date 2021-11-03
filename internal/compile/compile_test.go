package compile

import (
	"errors"
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/debug"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		debugger  *debug.DebuggerSet
		consumers []*Consumer
	}{
		{
			name:      "OK",
			debugger:  debug.NewSet(debug.None),
			consumers: []*Consumer{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := New(tc.debugger, tc.consumers...)

			assert.NotNil(t, c)
			assert.NotNil(t, c.parser)
			assert.Equal(t, tc.debugger, c.parser.debugger)
			assert.Equal(t, tc.consumers, c.parser.consumers)
		})
	}
}

func TestCompiler_Compile(t *testing.T) {
	tests := []struct {
		name          string
		consumers     []*Consumer
		path          string
		opts          ParseOptions
		expectedError string
	}{
		{
			name: "Success_SkipPackages",
			consumers: []*Consumer{
				{
					Name:    "tester",
					Package: func(*PackageInfo, *ast.Package) bool { return false },
				},
			},
			path: "./test/valid",
			opts: ParseOptions{
				SkipTestFiles: true,
			},
			expectedError: "",
		},
		{
			name: "Success_SkipFiles",
			consumers: []*Consumer{
				{
					Name:    "tester",
					Package: func(*PackageInfo, *ast.Package) bool { return true },
					FilePre: func(*FileInfo, *ast.File) bool { return false },
				},
			},
			path: "./test/valid",
			opts: ParseOptions{
				SkipTestFiles: true,
			},
			expectedError: "",
		},
		{
			name: "Success",
			consumers: []*Consumer{
				{
					Name:      "tester",
					Package:   func(*PackageInfo, *ast.Package) bool { return true },
					FilePre:   func(*FileInfo, *ast.File) bool { return true },
					Import:    func(*FileInfo, *ast.ImportSpec) {},
					Struct:    func(*TypeInfo, *ast.StructType) {},
					Interface: func(*TypeInfo, *ast.InterfaceType) {},
					FuncType:  func(*TypeInfo, *ast.FuncType) {},
					FuncDecl:  func(*FuncInfo, *ast.FuncType, *ast.BlockStmt) {},
					FilePost:  func(*FileInfo, *ast.File) error { return nil },
				},
			},
			path: "./test/valid",
			opts: ParseOptions{
				SkipTestFiles: true,
			},
			expectedError: "",
		},
		{
			name: "FilePostFails",
			consumers: []*Consumer{
				{
					Name:      "tester",
					Package:   func(*PackageInfo, *ast.Package) bool { return true },
					FilePre:   func(*FileInfo, *ast.File) bool { return true },
					Import:    func(*FileInfo, *ast.ImportSpec) {},
					Struct:    func(*TypeInfo, *ast.StructType) {},
					Interface: func(*TypeInfo, *ast.InterfaceType) {},
					FuncType:  func(*TypeInfo, *ast.FuncType) {},
					FuncDecl:  func(*FuncInfo, *ast.FuncType, *ast.BlockStmt) {},
					FilePost:  func(*FileInfo, *ast.File) error { return errors.New("file error") },
				},
			},
			path: "./test/valid",
			opts: ParseOptions{
				SkipTestFiles: true,
			},
			expectedError: "file error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			debugger := debug.NewSet(debug.None)
			c := New(debugger, tc.consumers...)

			err := c.Compile(tc.path, tc.opts)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
