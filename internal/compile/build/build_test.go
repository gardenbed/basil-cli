package build

import (
	"go/ast"
	"go/token"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/compile"
	"github.com/gardenbed/basil-cli/internal/ui"
)

func TestNew(t *testing.T) {
	ui := ui.NewInteractive(ui.Info)
	c := New(ui)

	assert.NotNil(t, c)
	assert.IsType(t, &compile.Compiler{}, c)
}

func TestBuilder_Package(t *testing.T) {
	tests := []struct {
		name             string
		info             *compile.Package
		pkg              *ast.Package
		expectedContinue bool
	}{
		{
			name: "OK",
			info: &compile.Package{},
			pkg: &ast.Package{
				Name: "lookup",
			},
			expectedContinue: true,
		},
		{
			name: "FilterMainPackage",
			info: &compile.Package{},
			pkg: &ast.Package{
				Name: "main",
			},
			expectedContinue: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := &builder{}

			cont := b.Package(tc.info, tc.pkg)

			assert.Equal(t, tc.expectedContinue, cont)
		})
	}
}

func TestBuilder_FilePre(t *testing.T) {
	tests := []struct {
		name             string
		info             *compile.File
		file             *ast.File
		expectedContinue bool
	}{
		{
			name:             "OK",
			info:             &compile.File{},
			file:             &ast.File{},
			expectedContinue: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := &builder{}

			cont := b.FilePre(tc.info, tc.file)

			assert.Equal(t, tc.expectedContinue, cont)
		})
	}
}

func TestBuilder_FilePost(t *testing.T) {
	tests := []struct {
		name          string
		imports       []ast.Spec
		decls         []ast.Decl
		info          *compile.File
		file          *ast.File
		expectedError string
	}{
		{
			name:          "NoDeclaration",
			imports:       nil,
			decls:         nil,
			info:          &compile.File{},
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
			info: &compile.File{
				Package: compile.Package{
					Module: compile.Module{
						Name: "github.com/octocat/service",
					},
					Name:        "lookup",
					ImportPath:  "github.com/octocat/service/internal/lookup",
					BaseDir:     "/dev/null",
					RelativeDir: "internal/lookup",
				},
				Name:    "lookup.go",
				FileSet: token.NewFileSet(),
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
			info: &compile.File{
				Package: compile.Package{
					Module: compile.Module{
						Name: "github.com/octocat/service",
					},
					Name:        "lookup",
					ImportPath:  "github.com/octocat/service/internal/lookup",
					BaseDir:     "./service",
					RelativeDir: "internal/lookup",
				},
				Name:    "lookup.go",
				FileSet: token.NewFileSet(),
			},
			file:          &ast.File{},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := &builder{
				imports: tc.imports,
				decls:   tc.decls,
			}

			err := b.FilePost(tc.info, tc.file)

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

func TestBuilder_Import(t *testing.T) {
	tests := []struct {
		name            string
		info            *compile.File
		spec            *ast.ImportSpec
		expectedImports []ast.Spec
	}{
		{
			name: "OK",
			info: &compile.File{},
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
			b := &builder{}

			b.Import(tc.info, tc.spec)

			assert.Equal(t, tc.expectedImports, b.imports)
		})
	}
}

func TestBuilder_Struct(t *testing.T) {
	tests := []struct {
		name          string
		info          *compile.Type
		node          *ast.StructType
		expectedDecls []ast.Decl
	}{
		{
			name: "Request",
			info: &compile.Type{
				File: compile.File{
					Package: compile.Package{
						Name: "lookup",
					},
				},
				Name: "Request",
			},
			node: &ast.StructType{
				Fields: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								&ast.Ident{Name: "ID"},
							},
							Type: &ast.Ident{Name: "string"},
						},
					},
				},
			},
			expectedDecls: []ast.Decl{
				// Type func
				&ast.FuncDecl{
					Name: &ast.Ident{Name: "Request"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "lookup"},
										Sel: &ast.Ident{Name: "Request"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.CallExpr{
												Fun: &ast.Ident{Name: "BuildRequest"},
											},
											Sel: &ast.Ident{Name: "Value"},
										},
									},
								},
							},
						},
					},
				},
				// Builder struct
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{Name: "RequestBuilder"},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "v"},
											},
											Type: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "lookup"},
												Sel: &ast.Ident{Name: "Request"},
											},
										},
									},
								},
							},
						},
					},
				},
				// Build func
				&ast.FuncDecl{
					Name: &ast.Ident{Name: "BuildRequest"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{Name: "RequestBuilder"},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CompositeLit{
										Type: &ast.Ident{Name: "RequestBuilder"},
										Elts: []ast.Expr{
											&ast.KeyValueExpr{
												Key: &ast.Ident{Name: "v"},
												Value: &ast.CompositeLit{
													Type: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "lookup"},
														Sel: &ast.Ident{Name: "Request"},
													},
													Elts: []ast.Expr{
														&ast.KeyValueExpr{
															Key: &ast.Ident{Name: "ID"},
															Value: &ast.CallExpr{
																Fun: &ast.SelectorExpr{
																	X:   &ast.Ident{Name: "value"},
																	Sel: &ast.Ident{Name: "String"},
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
				// Builder method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "RequestBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "WithID"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{
										{Name: "id"},
									},
									Type: &ast.Ident{Name: "string"},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{Name: "RequestBuilder"},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "b"},
											Sel: &ast.Ident{Name: "v"},
										},
										Sel: &ast.Ident{Name: "ID"},
									},
								},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{
									&ast.Ident{Name: "id"},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{Name: "b"},
								},
							},
						},
					},
				},
				// Value method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "RequestBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "Value"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "lookup"},
										Sel: &ast.Ident{Name: "Request"},
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
										X:   &ast.Ident{Name: "b"},
										Sel: &ast.Ident{Name: "v"},
									},
								},
							},
						},
					},
				},
				// Pointer method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "RequestBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "Pointer"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "lookup"},
											Sel: &ast.Ident{Name: "Request"},
										},
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
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "b"},
											Sel: &ast.Ident{Name: "v"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Response",
			info: &compile.Type{
				File: compile.File{
					Package: compile.Package{
						Name: "lookup",
					},
				},
				Name: "Response",
			},
			node: &ast.StructType{
				Fields: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								&ast.Ident{Name: "Name"},
							},
							Type: &ast.Ident{Name: "string"},
						},
					},
				},
			},
			expectedDecls: []ast.Decl{
				// Type func
				&ast.FuncDecl{
					Name: &ast.Ident{Name: "Response"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "lookup"},
										Sel: &ast.Ident{Name: "Response"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.CallExpr{
												Fun: &ast.Ident{Name: "BuildResponse"},
											},
											Sel: &ast.Ident{Name: "Value"},
										},
									},
								},
							},
						},
					},
				},
				// Builder struct
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{Name: "ResponseBuilder"},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "v"},
											},
											Type: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "lookup"},
												Sel: &ast.Ident{Name: "Response"},
											},
										},
									},
								},
							},
						},
					},
				},
				// Build func
				&ast.FuncDecl{
					Name: &ast.Ident{Name: "BuildResponse"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{Name: "ResponseBuilder"},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CompositeLit{
										Type: &ast.Ident{Name: "ResponseBuilder"},
										Elts: []ast.Expr{
											&ast.KeyValueExpr{
												Key: &ast.Ident{Name: "v"},
												Value: &ast.CompositeLit{
													Type: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "lookup"},
														Sel: &ast.Ident{Name: "Response"},
													},
													Elts: []ast.Expr{
														&ast.KeyValueExpr{
															Key: &ast.Ident{Name: "Name"},
															Value: &ast.CallExpr{
																Fun: &ast.SelectorExpr{
																	X:   &ast.Ident{Name: "value"},
																	Sel: &ast.Ident{Name: "String"},
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
				// Builder method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "ResponseBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "WithName"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{
										{Name: "name"},
									},
									Type: &ast.Ident{Name: "string"},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{Name: "ResponseBuilder"},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "b"},
											Sel: &ast.Ident{Name: "v"},
										},
										Sel: &ast.Ident{Name: "Name"},
									},
								},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{
									&ast.Ident{Name: "name"},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{Name: "b"},
								},
							},
						},
					},
				},
				// Value method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "ResponseBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "Value"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "lookup"},
										Sel: &ast.Ident{Name: "Response"},
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
										X:   &ast.Ident{Name: "b"},
										Sel: &ast.Ident{Name: "v"},
									},
								},
							},
						},
					},
				},
				// Pointer method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "ResponseBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "Pointer"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "lookup"},
											Sel: &ast.Ident{Name: "Response"},
										},
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
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "b"},
											Sel: &ast.Ident{Name: "v"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "EmbeddedStruct",
			info: &compile.Type{
				File: compile.File{
					Package: compile.Package{
						Name: "account",
					},
				},
				Name: "Account",
			},
			node: &ast.StructType{
				Fields: &ast.FieldList{
					List: []*ast.Field{
						{
							Type: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "common"},
								Sel: &ast.Ident{Name: "Address"},
							},
						},
					},
				},
			},
			expectedDecls: []ast.Decl{
				// Type func
				&ast.FuncDecl{
					Name: &ast.Ident{Name: "Account"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "account"},
										Sel: &ast.Ident{Name: "Account"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.CallExpr{
												Fun: &ast.Ident{Name: "BuildAccount"},
											},
											Sel: &ast.Ident{Name: "Value"},
										},
									},
								},
							},
						},
					},
				},
				// Builder struct
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{Name: "AccountBuilder"},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "v"},
											},
											Type: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "account"},
												Sel: &ast.Ident{Name: "Account"},
											},
										},
									},
								},
							},
						},
					},
				},
				// Build func
				&ast.FuncDecl{
					Name: &ast.Ident{Name: "BuildAccount"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{Name: "AccountBuilder"},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CompositeLit{
										Type: &ast.Ident{Name: "AccountBuilder"},
										Elts: []ast.Expr{
											&ast.KeyValueExpr{
												Key: &ast.Ident{Name: "v"},
												Value: &ast.CompositeLit{
													Type: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "account"},
														Sel: &ast.Ident{Name: "Account"},
													},
													Elts: []ast.Expr{
														&ast.KeyValueExpr{
															Key:   &ast.Ident{Name: "Address"},
															Value: &ast.Ident{Name: "nil"},
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
				// Builder method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "AccountBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "WithAddress"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{
										{Name: "address"},
									},
									Type: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "common"},
										Sel: &ast.Ident{Name: "Address"},
									},
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{Name: "AccountBuilder"},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "b"},
											Sel: &ast.Ident{Name: "v"},
										},
										Sel: &ast.Ident{Name: "Address"},
									},
								},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{
									&ast.Ident{Name: "address"},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{Name: "b"},
								},
							},
						},
					},
				},
				// Value method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "AccountBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "Value"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "account"},
										Sel: &ast.Ident{Name: "Account"},
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
										X:   &ast.Ident{Name: "b"},
										Sel: &ast.Ident{Name: "v"},
									},
								},
							},
						},
					},
				},
				// Pointer method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "AccountBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "Pointer"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "account"},
											Sel: &ast.Ident{Name: "Account"},
										},
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
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "b"},
											Sel: &ast.Ident{Name: "v"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "UnexportedField",
			info: &compile.Type{
				File: compile.File{
					Package: compile.Package{
						Name: "example",
					},
				},
				Name: "Example",
			},
			node: &ast.StructType{
				Fields: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								&ast.Ident{Name: "internal"},
							},
							Type: &ast.Ident{Name: "string"},
						},
					},
				},
			},
			expectedDecls: []ast.Decl{
				// Type func
				&ast.FuncDecl{
					Name: &ast.Ident{Name: "Example"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "example"},
										Sel: &ast.Ident{Name: "Example"},
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.CallExpr{
												Fun: &ast.Ident{Name: "BuildExample"},
											},
											Sel: &ast.Ident{Name: "Value"},
										},
									},
								},
							},
						},
					},
				},
				// Builder struct
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{Name: "ExampleBuilder"},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{Name: "v"},
											},
											Type: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "example"},
												Sel: &ast.Ident{Name: "Example"},
											},
										},
									},
								},
							},
						},
					},
				},
				// Build func
				&ast.FuncDecl{
					Name: &ast.Ident{Name: "BuildExample"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{Name: "ExampleBuilder"},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CompositeLit{
										Type: &ast.Ident{Name: "ExampleBuilder"},
										Elts: []ast.Expr{
											&ast.KeyValueExpr{
												Key: &ast.Ident{Name: "v"},
												Value: &ast.CompositeLit{
													Type: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "example"},
														Sel: &ast.Ident{Name: "Example"},
													},
													Elts: []ast.Expr{},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				// Builder method
				// Value method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "ExampleBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "Value"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "example"},
										Sel: &ast.Ident{Name: "Example"},
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
										X:   &ast.Ident{Name: "b"},
										Sel: &ast.Ident{Name: "v"},
									},
								},
							},
						},
					},
				},
				// Pointer method
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "b"},
								},
								Type: &ast.Ident{Name: "ExampleBuilder"},
							},
						},
					},
					Name: &ast.Ident{Name: "Pointer"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "example"},
											Sel: &ast.Ident{Name: "Example"},
										},
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
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "b"},
											Sel: &ast.Ident{Name: "v"},
										},
									},
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
			b := &builder{}

			b.Struct(tc.info, tc.node)

			assert.Equal(t, tc.expectedDecls, b.decls)
		})
	}
}
