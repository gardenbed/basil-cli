package compile

import (
	"errors"
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/ui"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		ui        ui.UI
		consumers []*Consumer
	}{
		{
			name:      "OK",
			ui:        ui.NewNop(),
			consumers: []*Consumer{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := New(tc.ui, tc.consumers...)

			assert.NotNil(t, c)
			assert.NotNil(t, c.parser)
			assert.Equal(t, tc.ui, c.parser.ui)
			assert.Equal(t, tc.consumers, c.parser.consumers)
		})
	}
}

func TestCompiler_Compile(t *testing.T) {
	tests := []struct {
		name          string
		consumers     []*Consumer
		packages      string
		opts          ParseOptions
		expectedError string
	}{
		{
			name: "Success_SkipPackages",
			consumers: []*Consumer{
				{
					Name:    "tester",
					Package: func(*Package, *ast.Package) bool { return false },
				},
			},
			packages: "./test/valid/...",
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
					Package: func(*Package, *ast.Package) bool { return true },
					FilePre: func(*File, *ast.File) bool { return false },
				},
			},
			packages: "./test/valid/...",
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
					Package:   func(*Package, *ast.Package) bool { return true },
					FilePre:   func(*File, *ast.File) bool { return true },
					Import:    func(*File, *ast.ImportSpec) {},
					Struct:    func(*Type, *ast.StructType) {},
					Interface: func(*Type, *ast.InterfaceType) {},
					FuncType:  func(*Type, *ast.FuncType) {},
					FuncDecl:  func(*Func, *ast.FuncType, *ast.BlockStmt) {},
					FilePost:  func(*File, *ast.File) error { return nil },
				},
			},
			packages: "./test/valid/...",
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
					Package:   func(*Package, *ast.Package) bool { return true },
					FilePre:   func(*File, *ast.File) bool { return true },
					Import:    func(*File, *ast.ImportSpec) {},
					Struct:    func(*Type, *ast.StructType) {},
					Interface: func(*Type, *ast.InterfaceType) {},
					FuncType:  func(*Type, *ast.FuncType) {},
					FuncDecl:  func(*Func, *ast.FuncType, *ast.BlockStmt) {},
					FilePost:  func(*File, *ast.File) error { return errors.New("file error") },
				},
			},
			packages: "./test/valid/...",
			opts: ParseOptions{
				SkipTestFiles: true,
			},
			expectedError: "file error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ui := ui.NewNop()
			c := New(ui, tc.consumers...)

			err := c.Compile(tc.packages, tc.opts)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
