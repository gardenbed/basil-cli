package compile

import (
	"errors"
	"go/ast"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/debug"
)

func TestTypeInfo_IsExported(t *testing.T) {
	tests := []struct {
		name               string
		info               *TypeInfo
		expectedIsExported bool
	}{
		{
			name: "Exported",
			info: &TypeInfo{
				TypeName: "Controller",
			},
			expectedIsExported: true,
		},
		{
			name: "Unexported",
			info: &TypeInfo{
				TypeName: "controller",
			},
			expectedIsExported: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isExported := tc.info.IsExported()

			assert.Equal(t, tc.expectedIsExported, isExported)
		})
	}
}

func TestFuncInfo_IsExported(t *testing.T) {
	tests := []struct {
		name               string
		info               *FuncInfo
		expectedIsExported bool
	}{
		{
			name: "Exported",
			info: &FuncInfo{
				FuncName: "Lookup",
			},
			expectedIsExported: true,
		},
		{
			name: "Unexported",
			info: &FuncInfo{
				FuncName: "lookup",
			},
			expectedIsExported: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isExported := tc.info.IsExported()

			assert.Equal(t, tc.expectedIsExported, isExported)
		})
	}
}

func TestFuncInfo_IsMethod(t *testing.T) {
	tests := []struct {
		name             string
		info             *FuncInfo
		expectedIsMethod bool
	}{
		{
			name:             "Function",
			info:             &FuncInfo{},
			expectedIsMethod: false,
		},
		{
			name: "Method",
			info: &FuncInfo{
				RecvName: "Lookup",
				RecvType: &ast.StarExpr{
					X: &ast.Ident{Name: "service"},
				},
			},
			expectedIsMethod: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isMethod := tc.info.IsMethod()

			assert.Equal(t, tc.expectedIsMethod, isMethod)
		})
	}
}

func TestParseOptions_MatchType(t *testing.T) {
	tests := []struct {
		name            string
		opts            ParseOptions
		typeName        *ast.Ident
		expectedMatched bool
	}{
		{
			name:            "Matched_NoFilter",
			opts:            ParseOptions{},
			typeName:        &ast.Ident{Name: "Request"},
			expectedMatched: true,
		},
		{
			name: "Matched_TypeName",
			opts: ParseOptions{
				TypeNames: []string{"Response"},
			},
			typeName:        &ast.Ident{Name: "Response"},
			expectedMatched: true,
		},
		{
			name: "Matched_TypeRegexp",
			opts: ParseOptions{
				TypeRegexp: regexp.MustCompile(`Service$`),
			},
			typeName:        &ast.Ident{Name: "ExampleService"},
			expectedMatched: true,
		},
		{
			name: "NotMatched",
			opts: ParseOptions{
				TypeNames:  []string{"Request", "Response"},
				TypeRegexp: regexp.MustCompile(`Service$`),
			},
			typeName:        &ast.Ident{Name: "Helper"},
			expectedMatched: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matched := tc.opts.matchType(tc.typeName)

			assert.Equal(t, tc.expectedMatched, matched)
		})
	}
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name          string
		consumers     []*Consumer
		path          string
		opts          ParseOptions
		expectedError string
	}{
		{
			name:          "PathNotExist",
			path:          "/foo",
			opts:          ParseOptions{},
			expectedError: "stat /foo: no such file or directory",
		},
		{
			name:          "PathNotDirectory",
			path:          "/dev/null",
			opts:          ParseOptions{},
			expectedError: "stat /dev/null/go.mod: not a directory",
		},
		{
			name:          "InvalidModule",
			path:          "./test/invalid_module",
			opts:          ParseOptions{},
			expectedError: "invalid go.mod file: no module name found",
		},
		{
			name:          "InvalidCode",
			path:          "./test/invalid_code",
			opts:          ParseOptions{},
			expectedError: "test/invalid_code/main.go:3:11: expected 'STRING', found newline (and 1 more errors)",
		},
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
			path:          "./test/valid",
			opts:          ParseOptions{},
			expectedError: "file error",
		},
		{
			name: "FilePostFails_MergePackageFiles",
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
				MergePackageFiles: true,
			},
			expectedError: "file error",
		},
		{
			name: "Success_MergePackageFiles",
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
				MergePackageFiles: true,
			},
			expectedError: "",
		},
		{
			name: "Success_SkipTestFiles",
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
			path:          "./test/valid",
			opts:          ParseOptions{},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &parser{
				debugger:  debug.NewSet(debug.None),
				consumers: tc.consumers,
			}

			err := p.Parse(tc.path, tc.opts)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
