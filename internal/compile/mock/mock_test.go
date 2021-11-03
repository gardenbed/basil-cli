package mock

import (
	"go/ast"
	"go/token"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/compile"
	"github.com/gardenbed/basil-cli/internal/debug"
)

func TestNew(t *testing.T) {
	c := New(debug.Info)

	assert.NotNil(t, c)
	assert.IsType(t, &compile.Compiler{}, c)
}

func TestMocker_Package(t *testing.T) {
	tests := []struct {
		name             string
		info             *compile.PackageInfo
		pkg              *ast.Package
		expectedContinue bool
	}{
		{
			name: "OK",
			info: &compile.PackageInfo{},
			pkg: &ast.Package{
				Name: "lookup",
			},
			expectedContinue: true,
		},
		{
			name: "FilterMainPackage",
			info: &compile.PackageInfo{},
			pkg: &ast.Package{
				Name: "main",
			},
			expectedContinue: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mocker{}

			cont := m.Package(tc.info, tc.pkg)

			assert.Equal(t, tc.expectedContinue, cont)
		})
	}
}

func TestMocker_FilePre(t *testing.T) {
	tests := []struct {
		name             string
		info             *compile.FileInfo
		file             *ast.File
		expectedContinue bool
	}{
		{
			name:             "OK",
			info:             &compile.FileInfo{},
			file:             &ast.File{},
			expectedContinue: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mocker{}

			cont := m.FilePre(tc.info, tc.file)

			assert.Equal(t, tc.expectedContinue, cont)
		})
	}
}

func TestMocker_FilePost(t *testing.T) {
	tests := []struct {
		name          string
		imports       []ast.Spec
		decls         []ast.Decl
		info          *compile.FileInfo
		file          *ast.File
		expectedError string
	}{
		{
			name:          "NoDeclaration",
			imports:       nil,
			decls:         nil,
			info:          &compile.FileInfo{},
			file:          &ast.File{},
			expectedError: "",
		},
		{
			name: "WriteFileFails",
			imports: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{Value: `"fmt"`},
				},
			},
			decls: []ast.Decl{
				&ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{
								&ast.Ident{Name: "dummy"},
							},
							Type: &ast.Ident{Name: "string"},
						},
					},
				},
			},
			info: &compile.FileInfo{
				PackageInfo: compile.PackageInfo{
					ModuleName:  "github.com/octocat/service",
					PackageName: "lookup",
					ImportPath:  "github.com/octocat/service/internal/lookup",
					BaseDir:     "/dev/null",
					RelativeDir: "internal/lookup",
				},
				FileName: "lookup.go",
				FileSet:  token.NewFileSet(),
			},
			file:          &ast.File{},
			expectedError: "mkdir /dev/null: not a directory",
		},
		{
			name: "Success",
			imports: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{Value: `"fmt"`},
				},
			},
			decls: []ast.Decl{
				&ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{
								&ast.Ident{Name: "dummy"},
							},
							Type: &ast.Ident{Name: "string"},
						},
					},
				},
			},
			info: &compile.FileInfo{
				PackageInfo: compile.PackageInfo{
					ModuleName:  "github.com/octocat/service",
					PackageName: "lookup",
					ImportPath:  "github.com/octocat/service/internal/lookup",
					BaseDir:     "./service",
					RelativeDir: "internal/lookup",
				},
				FileName: "lookup.go",
				FileSet:  token.NewFileSet(),
			},
			file:          &ast.File{},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mocker{
				imports: tc.imports,
				decls:   tc.decls,
			}

			err := m.FilePost(tc.info, tc.file)

			// Cleanup
			defer os.RemoveAll("./service")

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMocker_Import(t *testing.T) {
	tests := []struct {
		name            string
		info            *compile.FileInfo
		spec            *ast.ImportSpec
		expectedImports []ast.Spec
	}{
		{
			name: "OK",
			info: &compile.FileInfo{},
			spec: &ast.ImportSpec{
				Path: &ast.BasicLit{Value: `"fmt"`},
			},
			expectedImports: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{Value: `"fmt"`},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mocker{}

			m.Import(tc.info, tc.spec)

			assert.Equal(t, tc.expectedImports, m.imports)
		})
	}
}

func TestMocker_Interface(t *testing.T) {
	tests := []struct {
		name          string
		info          *compile.TypeInfo
		node          *ast.InterfaceType
		expectedDecls []ast.Decl
	}{
		{
			name: "Service",
			info: &compile.TypeInfo{
				FileInfo: compile.FileInfo{
					PackageInfo: compile.PackageInfo{
						PackageName: "lookup",
					},
				},
				TypeName: "Service",
			},
			node: &ast.InterfaceType{
				Methods: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								&ast.Ident{Name: "Lookup"},
							},
							Type: &ast.FuncType{
								Params: &ast.FieldList{
									List: []*ast.Field{
										{
											Type: &ast.StarExpr{
												X: &ast.Ident{Name: "Request"},
											},
										},
									},
								},
								Results: &ast.FieldList{
									List: []*ast.Field{
										{
											Type: &ast.StarExpr{
												X: &ast.Ident{Name: "Response"},
											},
										},
										{
											Type: &ast.Ident{Name: "error"},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDecls: []ast.Decl{
				// Mocker struct
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: "ServiceMocker",
							},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "t"},
											},
											Type: &ast.StarExpr{
												X: &ast.SelectorExpr{
													X:   &ast.Ident{Name: "testing"},
													Sel: &ast.Ident{Name: "T"},
												},
											},
										},
										{
											Names: []*ast.Ident{
												{Name: "spew"},
											},
											Type: &ast.StarExpr{
												X: &ast.SelectorExpr{
													X:   &ast.Ident{Name: "spew"},
													Sel: &ast.Ident{Name: "ConfigState"},
												},
											},
										},
										{
											Names: []*ast.Ident{
												{Name: "expectations"},
											},
											Type: &ast.StarExpr{
												X: &ast.Ident{Name: "ServiceExpectations"},
											},
										},
									},
								},
							},
						},
					},
				},
				// Mock func
				&ast.FuncDecl{
					Name: &ast.Ident{
						Name: "MockService",
					},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{
										{Name: "t"},
									},
									Type: &ast.StarExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "testing"},
											Sel: &ast.Ident{Name: "T"},
										},
									},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.Ident{Name: "ServiceMocker"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.UnaryExpr{
										Op: token.AND,
										X: &ast.CompositeLit{
											Type: &ast.Ident{Name: "ServiceMocker"},
											Elts: []ast.Expr{
												&ast.KeyValueExpr{
													Key:   &ast.Ident{Name: "t"},
													Value: &ast.Ident{Name: "t"},
												},
												&ast.KeyValueExpr{
													Key: &ast.Ident{Name: "spew"},
													Value: &ast.UnaryExpr{
														Op: token.AND,
														X: &ast.CompositeLit{
															Type: &ast.SelectorExpr{
																X:   &ast.Ident{Name: "spew"},
																Sel: &ast.Ident{Name: "ConfigState"},
															},
															Elts: []ast.Expr{
																&ast.KeyValueExpr{
																	Key:   &ast.Ident{Name: "Indent"},
																	Value: &ast.BasicLit{Value: `"  "`},
																},
																&ast.KeyValueExpr{
																	Key:   &ast.Ident{Name: "DisablePointerAddresses"},
																	Value: &ast.Ident{Name: "true"},
																},
																&ast.KeyValueExpr{
																	Key:   &ast.Ident{Name: "DisableCapacities"},
																	Value: &ast.Ident{Name: "true"},
																},
																&ast.KeyValueExpr{
																	Key:   &ast.Ident{Name: "SortKeys"},
																	Value: &ast.Ident{Name: "true"},
																},
															},
														},
													},
												},
												&ast.KeyValueExpr{
													Key: &ast.Ident{Name: "expectations"},
													Value: &ast.CallExpr{
														Fun: &ast.Ident{Name: "new"},
														Args: []ast.Expr{
															&ast.Ident{Name: "ServiceExpectations"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				// Mocker Expect method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "m"},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{Name: "ServiceMocker"},
								},
							},
						},
					},
					Name: &ast.Ident{
						Name: "Expect",
					},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.Ident{Name: "ServiceExpectations"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.SelectorExpr{
										X:   &ast.Ident{Name: "m"},
										Sel: &ast.Ident{Name: "expectations"},
									},
								},
							},
						},
					},
				},
				// Mocker Impl method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "m"},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{Name: "ServiceMocker"},
								},
							},
						},
					},
					Name: &ast.Ident{
						Name: "Impl",
					},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "lookup"},
										Sel: &ast.Ident{Name: "Service"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.UnaryExpr{
										Op: token.AND,
										X: &ast.CompositeLit{
											Type: &ast.Ident{Name: "ServiceImpl"},
											Elts: []ast.Expr{
												&ast.KeyValueExpr{
													Key: &ast.Ident{Name: "t"},
													Value: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "m"},
														Sel: &ast.Ident{Name: "t"},
													},
												},
												&ast.KeyValueExpr{
													Key: &ast.Ident{Name: "spew"},
													Value: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "m"},
														Sel: &ast.Ident{Name: "spew"},
													},
												},
												&ast.KeyValueExpr{
													Key: &ast.Ident{Name: "expectations"},
													Value: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "m"},
														Sel: &ast.Ident{Name: "expectations"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				// Mocker Assert method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "m"},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{Name: "ServiceMocker"},
								},
							},
						},
					},
					Name: &ast.Ident{
						Name: "Assert",
					},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.RangeStmt{
								Key:   &ast.Ident{Name: "_"},
								Value: &ast.Ident{Name: "e"},
								Tok:   token.DEFINE,
								X: &ast.SelectorExpr{
									X: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "m"},
										Sel: &ast.Ident{Name: "expectations"},
									},
									Sel: &ast.Ident{Name: "lookupExpectations"},
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.IfStmt{
											Cond: &ast.BinaryExpr{
												X: &ast.SelectorExpr{
													X:   &ast.Ident{Name: "e"},
													Sel: &ast.Ident{Name: "recorded"},
												},
												Op: token.EQL,
												Y:  &ast.Ident{Name: "nil"},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.ExprStmt{
														X: &ast.CallExpr{
															Fun: &ast.SelectorExpr{
																X: &ast.SelectorExpr{
																	X:   &ast.Ident{Name: "m"},
																	Sel: &ast.Ident{Name: "t"},
																},
																Sel: &ast.Ident{Name: "Errorf"},
															},
															Args: []ast.Expr{
																&ast.BasicLit{
																	Value: `"\nExpected Lookup method be called with %s"`,
																},
																&ast.CallExpr{
																	Fun: &ast.SelectorExpr{
																		X: &ast.SelectorExpr{
																			X:   &ast.Ident{Name: "m"},
																			Sel: &ast.Ident{Name: "spew"},
																		},
																		Sel: &ast.Ident{Name: "Sdump"},
																	},
																	Args: []ast.Expr{
																		&ast.SelectorExpr{
																			X:   &ast.Ident{Name: "e"},
																			Sel: &ast.Ident{Name: "inputs"},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				// Expectations struct
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: "ServiceExpectations",
							},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "lookupExpectations"},
											},
											Type: &ast.ArrayType{
												Elt: &ast.StarExpr{
													X: &ast.Ident{Name: "LookupExpectation"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				// Expectations methods
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "e"},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{Name: "ServiceExpectations"},
								},
							},
						},
					},
					Name: &ast.Ident{Name: "Lookup"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.Ident{Name: "LookupExpectation"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.Ident{Name: "expectation"},
								},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.Ident{Name: "new"},
										Args: []ast.Expr{
											&ast.Ident{Name: "LookupExpectation"},
										},
									},
								},
							},
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.SelectorExpr{
										X:   &ast.Ident{Name: "e"},
										Sel: &ast.Ident{Name: "lookupExpectations"},
									},
								},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.Ident{Name: "append"},
										Args: []ast.Expr{
											&ast.SelectorExpr{
												X:   &ast.Ident{Name: "e"},
												Sel: &ast.Ident{Name: "lookupExpectations"},
											},
											&ast.Ident{Name: "expectation"},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{Name: "expectation"},
								},
							},
						},
					},
				},
				// Expectation structs
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: "LookupExpectation",
							},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "inputs"},
											},
											Type: &ast.StarExpr{
												X: &ast.Ident{Name: "lookupInputs"},
											},
										},
										{
											Names: []*ast.Ident{
												{Name: "outputs"},
											},
											Type: &ast.StarExpr{
												X: &ast.Ident{Name: "lookupOutputs"},
											},
										},
										{
											Names: []*ast.Ident{
												{Name: "callback"},
											},
											Type: &ast.FuncType{
												Params: &ast.FieldList{
													List: []*ast.Field{
														{
															Type: &ast.StarExpr{
																X: &ast.Ident{Name: "Request"},
															},
														},
													},
												},
												Results: &ast.FieldList{
													List: []*ast.Field{
														{
															Type: &ast.StarExpr{
																X: &ast.Ident{Name: "Response"},
															},
														},
														{
															Type: &ast.Ident{Name: "error"},
														},
													},
												},
											},
										},
										{
											Names: []*ast.Ident{
												{Name: "recorded"},
											},
											Type: &ast.StarExpr{
												X: &ast.Ident{Name: "lookupInputs"},
											},
										},
									},
								},
							},
						},
					},
				},
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: "lookupInputs",
							},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "request"},
											},
											Type: &ast.StarExpr{
												X: &ast.Ident{Name: "Request"},
											},
										},
									},
								},
							},
						},
					},
				},
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: "lookupOutputs",
							},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "response"},
											},
											Type: &ast.StarExpr{
												X: &ast.Ident{Name: "Response"},
											},
										},
										{
											Names: []*ast.Ident{
												{Name: "error"},
											},
											Type: &ast.Ident{Name: "error"},
										},
									},
								},
							},
						},
					},
				},
				// Expectation WithArgs method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "e"},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{Name: "LookupExpectation"},
								},
							},
						},
					},
					Name: &ast.Ident{Name: "WithArgs"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{
										{Name: "request"},
									},
									Type: &ast.StarExpr{
										X: &ast.Ident{Name: "Request"},
									},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.Ident{Name: "LookupExpectation"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.SelectorExpr{
										X:   &ast.Ident{Name: "e"},
										Sel: &ast.Ident{Name: "inputs"},
									},
								},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{
									&ast.UnaryExpr{
										Op: token.AND,
										X: &ast.CompositeLit{
											Type: &ast.Ident{Name: "lookupInputs"},
											Elts: []ast.Expr{
												&ast.KeyValueExpr{
													Key:   &ast.Ident{Name: "request"},
													Value: &ast.Ident{Name: "request"},
												},
											},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{Name: "e"},
								},
							},
						},
					},
				},
				// Expectation Return method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "e"},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{Name: "LookupExpectation"},
								},
							},
						},
					},
					Name: &ast.Ident{Name: "Return"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{
										{Name: "response"},
									},
									Type: &ast.StarExpr{
										X: &ast.Ident{Name: "Response"},
									},
								},
								{
									Names: []*ast.Ident{
										{Name: "error"},
									},
									Type: &ast.Ident{Name: "error"},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.Ident{Name: "LookupExpectation"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.SelectorExpr{
										X:   &ast.Ident{Name: "e"},
										Sel: &ast.Ident{Name: "outputs"},
									},
								},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{
									&ast.UnaryExpr{
										Op: token.AND,
										X: &ast.CompositeLit{
											Type: &ast.Ident{Name: "lookupOutputs"},
											Elts: []ast.Expr{
												&ast.KeyValueExpr{
													Key:   &ast.Ident{Name: "response"},
													Value: &ast.Ident{Name: "response"},
												},
												&ast.KeyValueExpr{
													Key:   &ast.Ident{Name: "error"},
													Value: &ast.Ident{Name: "error"},
												},
											},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{Name: "e"},
								},
							},
						},
					},
				},
				// Expectation Call method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "e"},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{Name: "LookupExpectation"},
								},
							},
						},
					},
					Name: &ast.Ident{Name: "Call"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{
										{Name: "callback"},
									},
									Type: &ast.FuncType{
										Params: &ast.FieldList{
											List: []*ast.Field{
												{
													Type: &ast.StarExpr{
														X: &ast.Ident{Name: "Request"},
													},
												},
											},
										},
										Results: &ast.FieldList{
											List: []*ast.Field{
												{
													Type: &ast.StarExpr{
														X: &ast.Ident{Name: "Response"},
													},
												},
												{
													Type: &ast.Ident{Name: "error"},
												},
											},
										},
									},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.Ident{Name: "LookupExpectation"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.SelectorExpr{
										X:   &ast.Ident{Name: "e"},
										Sel: &ast.Ident{Name: "callback"},
									},
								},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{
									&ast.Ident{Name: "callback"},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{Name: "e"},
								},
							},
						},
					},
				},
				// Implementation struct
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: "ServiceImpl",
							},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "t"},
											},
											Type: &ast.StarExpr{
												X: &ast.SelectorExpr{
													X:   &ast.Ident{Name: "testing"},
													Sel: &ast.Ident{Name: "T"},
												},
											},
										},
										{
											Names: []*ast.Ident{
												{Name: "spew"},
											},
											Type: &ast.StarExpr{
												X: &ast.SelectorExpr{
													X:   &ast.Ident{Name: "spew"},
													Sel: &ast.Ident{Name: "ConfigState"},
												},
											},
										},
										{
											Names: []*ast.Ident{
												{Name: "expectations"},
											},
											Type: &ast.StarExpr{
												X: &ast.Ident{Name: "ServiceExpectations"},
											},
										},
									},
								},
							},
						},
					},
				},
				// Implementation methods
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "i"},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{Name: "ServiceImpl"},
								},
							},
						},
					},
					Name: &ast.Ident{Name: "Lookup"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{
										{Name: "request"},
									},
									Type: &ast.StarExpr{
										X: &ast.Ident{Name: "Request"},
									},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.Ident{Name: "Response"},
									},
								},
								{
									Type: &ast.Ident{Name: "error"},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.Ident{Name: "inputs"},
								},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{
									&ast.UnaryExpr{
										Op: token.AND,
										X: &ast.CompositeLit{
											Type: &ast.Ident{Name: "lookupInputs"},
											Elts: []ast.Expr{
												&ast.KeyValueExpr{
													Key:   &ast.Ident{Name: "request"},
													Value: &ast.Ident{Name: "request"},
												},
											},
										},
									},
								},
							},
							&ast.RangeStmt{
								Key:   &ast.Ident{Name: "_"},
								Value: &ast.Ident{Name: "e"},
								Tok:   token.DEFINE,
								X: &ast.SelectorExpr{
									X: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "i"},
										Sel: &ast.Ident{Name: "expectations"},
									},
									Sel: &ast.Ident{Name: "lookupExpectations"},
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.IfStmt{
											Cond: &ast.BinaryExpr{
												X: &ast.BinaryExpr{
													X: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "e"},
														Sel: &ast.Ident{Name: "inputs"},
													},
													Op: token.EQL,
													Y:  &ast.Ident{Name: "nil"},
												},
												Op: token.LOR,
												Y: &ast.CallExpr{
													Fun: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "reflect"},
														Sel: &ast.Ident{Name: "DeepEqual"},
													},
													Args: []ast.Expr{
														&ast.SelectorExpr{
															X:   &ast.Ident{Name: "e"},
															Sel: &ast.Ident{Name: "inputs"},
														},
														&ast.Ident{Name: "inputs"},
													},
												},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.AssignStmt{
														Lhs: []ast.Expr{
															&ast.SelectorExpr{
																X:   &ast.Ident{Name: "e"},
																Sel: &ast.Ident{Name: "recorded"},
															},
														},
														Tok: token.ASSIGN,
														Rhs: []ast.Expr{
															&ast.Ident{Name: "inputs"},
														},
													},
													&ast.IfStmt{
														Cond: &ast.BinaryExpr{
															X: &ast.SelectorExpr{
																X:   &ast.Ident{Name: "e"},
																Sel: &ast.Ident{Name: "callback"},
															},
															Op: token.NEQ,
															Y:  &ast.Ident{Name: "nil"},
														},
														Body: &ast.BlockStmt{
															List: []ast.Stmt{
																&ast.ReturnStmt{
																	Results: []ast.Expr{
																		&ast.CallExpr{
																			Fun: &ast.SelectorExpr{
																				X:   &ast.Ident{Name: "e"},
																				Sel: &ast.Ident{Name: "callback"},
																			},
																			Args: []ast.Expr{
																				&ast.Ident{Name: "request"},
																			},
																		},
																	},
																},
															},
														},
													},
													&ast.ReturnStmt{
														Results: []ast.Expr{
															&ast.SelectorExpr{
																X: &ast.SelectorExpr{
																	X:   &ast.Ident{Name: "e"},
																	Sel: &ast.Ident{Name: "outputs"},
																},
																Sel: &ast.Ident{Name: "response"},
															},
															&ast.SelectorExpr{
																X: &ast.SelectorExpr{
																	X:   &ast.Ident{Name: "e"},
																	Sel: &ast.Ident{Name: "outputs"},
																},
																Sel: &ast.Ident{Name: "error"},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "i"},
											Sel: &ast.Ident{Name: "t"},
										},
										Sel: &ast.Ident{Name: "Errorf"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Value: `"\nExpectation missing: Lookup method called with %s"`},
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X: &ast.SelectorExpr{
													X:   &ast.Ident{Name: "i"},
													Sel: &ast.Ident{Name: "spew"},
												},
												Sel: &ast.Ident{Name: "Sdump"},
											},
											Args: []ast.Expr{
												&ast.Ident{Name: "inputs"},
											},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{Name: "nil"},
									&ast.Ident{Name: "nil"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mocker{}

			m.Interface(tc.info, tc.node)

			assert.Equal(t, tc.expectedDecls, m.decls)
		})
	}
}
