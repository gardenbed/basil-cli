package mocker

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmbeddedInterface(t *testing.T) {
	tests := []struct {
		name     string
		method   *ast.Field
		expected bool
	}{
		{
			name: "EmbeddedInterface",
			method: &ast.Field{
				Type: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "http"},
					Sel: &ast.Ident{Name: "Handler"},
				},
			},
			expected: true,
		},
		{
			name: "Method",
			method: &ast.Field{
				Names: []*ast.Ident{
					{Name: "Lookup"},
				},
				Type: &ast.FuncType{},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, isEmbeddedInterface(tc.method))
		})
	}
}

func TestIsMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   *ast.Field
		expected bool
	}{
		{
			name: "Method",
			method: &ast.Field{
				Names: []*ast.Ident{
					{Name: "Lookup"},
				},
				Type: &ast.FuncType{},
			},
			expected: true,
		},
		{
			name: "EmbeddedInterface",
			method: &ast.Field{
				Type: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "http"},
					Sel: &ast.Ident{Name: "Handler"},
				},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, isMethod(tc.method))
		})
	}
}

func TestNormalizeFieldList(t *testing.T) {
	tests := []struct {
		name              string
		fieldList         *ast.FieldList
		expectedFieldList *ast.FieldList
	}{
		{
			name:              "NilFieldList",
			fieldList:         nil,
			expectedFieldList: &ast.FieldList{},
		},
		{
			// s string
			name: "NamedFields",
			fieldList: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{Name: "s"},
						},
						Type: &ast.Ident{Name: "string"},
					},
				},
			},
			expectedFieldList: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{Name: "s"},
						},
						Type: &ast.Ident{Name: "string"},
					},
				},
			},
		},
		{
			// string
			name: "UnnamedFields",
			fieldList: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.Ident{Name: "string"},
					},
				},
			},
			expectedFieldList: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{Name: "string"},
						},
						Type: &ast.Ident{Name: "string"},
					},
				},
			},
		},
		{
			// strings ...string
			name: "TrailingFields",
			fieldList: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{Name: "strings"},
						},
						Type: &ast.Ellipsis{
							Elt: &ast.Ident{Name: "string"},
						},
					},
				},
			},
			expectedFieldList: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{Name: "strings"},
						},
						Type: &ast.ArrayType{
							Elt: &ast.Ident{Name: "string"},
						},
					},
				},
			},
		},
		{
			// ..string
			name: "UnnamedTrailingFields",
			fieldList: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.Ellipsis{
							Elt: &ast.Ident{Name: "string"},
						},
					},
				},
			},
			expectedFieldList: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{Name: "string"},
						},
						Type: &ast.ArrayType{
							Elt: &ast.Ident{Name: "string"},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fieldList := normalizeFieldList(tc.fieldList)

			assert.Equal(t, tc.expectedFieldList, fieldList)
		})
	}
}

func TestCreateKeyValueExprList(t *testing.T) {
	tests := []struct {
		name          string
		fieldList     *ast.FieldList
		expectedExprs []ast.Expr
	}{
		{
			name:          "NilFieldList",
			fieldList:     nil,
			expectedExprs: []ast.Expr{},
		},
		{
			// (*Request) --> { request: request }
			name: "Params",
			fieldList: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{
							X: &ast.Ident{Name: "Request"},
						},
					},
				},
			},
			expectedExprs: []ast.Expr{
				&ast.KeyValueExpr{
					Key:   &ast.Ident{Name: "request"},
					Value: &ast.Ident{Name: "request"},
				},
			},
		},
		{
			// (*Response, error) --> { response: response, error: error }
			name: "Results",
			fieldList: &ast.FieldList{
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
			expectedExprs: []ast.Expr{
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exprs := createKeyValueExprList(tc.fieldList)

			assert.Equal(t, tc.expectedExprs, exprs)
		})
	}
}

func TestCreateZeroValueExpr(t *testing.T) {
	tests := []struct {
		name         string
		typ          ast.Expr
		expectedExpr ast.Expr
	}{
		{
			name:         "error",
			typ:          &ast.Ident{Name: "error"},
			expectedExpr: &ast.Ident{Name: "nil"},
		},
		{
			name:         "bool",
			typ:          &ast.Ident{Name: "bool"},
			expectedExpr: &ast.Ident{Name: "false"},
		},
		{
			name:         "string",
			typ:          &ast.Ident{Name: "string"},
			expectedExpr: &ast.BasicLit{Kind: token.STRING, Value: `""`},
		},
		{
			name:         "byte",
			typ:          &ast.Ident{Name: "byte"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "rune",
			typ:          &ast.Ident{Name: "rune"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "int",
			typ:          &ast.Ident{Name: "int"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "int8",
			typ:          &ast.Ident{Name: "int8"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "int16",
			typ:          &ast.Ident{Name: "int16"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "int32",
			typ:          &ast.Ident{Name: "int32"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "int64",
			typ:          &ast.Ident{Name: "int64"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "uint",
			typ:          &ast.Ident{Name: "int"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "uint8",
			typ:          &ast.Ident{Name: "int8"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "uint16",
			typ:          &ast.Ident{Name: "int16"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "uint32",
			typ:          &ast.Ident{Name: "int32"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "uint64",
			typ:          &ast.Ident{Name: "int64"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "uintptr",
			typ:          &ast.Ident{Name: "uintptr"},
			expectedExpr: &ast.BasicLit{Kind: token.INT, Value: "0"},
		},
		{
			name:         "float32",
			typ:          &ast.Ident{Name: "float32"},
			expectedExpr: &ast.BasicLit{Kind: token.FLOAT, Value: "0.0"},
		},
		{
			name:         "float64",
			typ:          &ast.Ident{Name: "float64"},
			expectedExpr: &ast.BasicLit{Kind: token.FLOAT, Value: "0.0"},
		},
		{
			name:         "complex64",
			typ:          &ast.Ident{Name: "complex64"},
			expectedExpr: &ast.BasicLit{Kind: token.IMAG, Value: "0.0i"},
		},
		{
			name:         "complex128",
			typ:          &ast.Ident{Name: "complex128"},
			expectedExpr: &ast.BasicLit{Kind: token.IMAG, Value: "0.0i"},
		},
		{
			// Address --> Address{}
			name: "Struct_SamePackage",
			typ:  &ast.Ident{Name: "Address"},
			expectedExpr: &ast.CompositeLit{
				Type: &ast.Ident{Name: "Address"},
			},
		},
		{
			// http.Transport --> http.Transport{}
			name: "Struct_OtherPackage",
			typ: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "http"},
				Sel: &ast.Ident{Name: "Transport"},
			},
			expectedExpr: &ast.CompositeLit{
				Type: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "http"},
					Sel: &ast.Ident{Name: "Transport"},
				},
			},
		},
		{
			// *int --> nil
			name: "Pointer",
			typ: &ast.StarExpr{
				X: &ast.Ident{Name: "int"},
			},
			expectedExpr: &ast.Ident{Name: "nil"},
		},
		{
			// []int --> nil
			name: "Slice",
			typ: &ast.ArrayType{
				Elt: &ast.Ident{Name: "int"},
			},
			expectedExpr: &ast.Ident{Name: "nil"},
		},
		{
			// map[int]string --> nil
			name: "Map",
			typ: &ast.MapType{
				Key:   &ast.Ident{Name: "int"},
				Value: &ast.Ident{Name: "string"},
			},
			expectedExpr: &ast.Ident{Name: "nil"},
		},
		{
			// chan error --> nil
			name: "Channel",
			typ: &ast.ChanType{
				Value: &ast.Ident{Name: "error"},
			},
			expectedExpr: &ast.Ident{Name: "nil"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expr := createZeroValueExpr(tc.typ)

			assert.Equal(t, tc.expectedExpr, expr)
		})
	}
}
