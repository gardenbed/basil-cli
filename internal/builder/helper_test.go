package builder

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFieldInitExpr(t *testing.T) {
	tests := []struct {
		name         string
		id           *ast.Ident
		typ          ast.Expr
		expectedExpr *ast.KeyValueExpr
	}{
		{
			name: "error",
			id:   &ast.Ident{Name: "e"},
			typ:  &ast.Ident{Name: "error"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "e"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Error"},
					},
				},
			},
		},
		{
			name: "bool",
			id:   &ast.Ident{Name: "b"},
			typ:  &ast.Ident{Name: "bool"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "b"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Bool"},
					},
				},
			},
		},
		{
			name: "string",
			id:   &ast.Ident{Name: "s"},
			typ:  &ast.Ident{Name: "string"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "s"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "String"},
					},
				},
			},
		},
		{
			name: "byte",
			id:   &ast.Ident{Name: "b"},
			typ:  &ast.Ident{Name: "byte"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "b"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Byte"},
					},
				},
			},
		},
		{
			name: "rune",
			id:   &ast.Ident{Name: "r"},
			typ:  &ast.Ident{Name: "rune"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "r"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Rune"},
					},
				},
			},
		},
		{
			name: "int",
			id:   &ast.Ident{Name: "i"},
			typ:  &ast.Ident{Name: "int"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "i"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Int"},
					},
				},
			},
		},
		{
			name: "int8",
			id:   &ast.Ident{Name: "i"},
			typ:  &ast.Ident{Name: "int8"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "i"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Int8"},
					},
				},
			},
		},
		{
			name: "int16",
			id:   &ast.Ident{Name: "i"},
			typ:  &ast.Ident{Name: "int16"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "i"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Int16"},
					},
				},
			},
		},
		{
			name: "int32",
			id:   &ast.Ident{Name: "i"},
			typ:  &ast.Ident{Name: "int32"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "i"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Int32"},
					},
				},
			},
		},
		{
			name: "int64",
			id:   &ast.Ident{Name: "i"},
			typ:  &ast.Ident{Name: "int64"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "i"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Int64"},
					},
				},
			},
		},
		{
			name: "uint",
			id:   &ast.Ident{Name: "u"},
			typ:  &ast.Ident{Name: "uint"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "u"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Uint"},
					},
				},
			},
		},
		{
			name: "uint8",
			id:   &ast.Ident{Name: "u"},
			typ:  &ast.Ident{Name: "uint8"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "u"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Uint8"},
					},
				},
			},
		},
		{
			name: "uint16",
			id:   &ast.Ident{Name: "u"},
			typ:  &ast.Ident{Name: "uint16"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "u"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Uint16"},
					},
				},
			},
		},
		{
			name: "uint32",
			id:   &ast.Ident{Name: "u"},
			typ:  &ast.Ident{Name: "uint32"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "u"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Uint32"},
					},
				},
			},
		},
		{
			name: "uint64",
			id:   &ast.Ident{Name: "u"},
			typ:  &ast.Ident{Name: "uint64"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "u"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Uint64"},
					},
				},
			},
		},
		{
			name: "uintptr",
			id:   &ast.Ident{Name: "u"},
			typ:  &ast.Ident{Name: "uintptr"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "u"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Uintptr"},
					},
				},
			},
		},
		{
			name: "float32",
			id:   &ast.Ident{Name: "f"},
			typ:  &ast.Ident{Name: "float32"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "f"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Float32"},
					},
				},
			},
		},
		{
			name: "float64",
			id:   &ast.Ident{Name: "f"},
			typ:  &ast.Ident{Name: "float64"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "f"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Float64"},
					},
				},
			},
		},
		{
			name: "complex64",
			id:   &ast.Ident{Name: "c"},
			typ:  &ast.Ident{Name: "complex64"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "c"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Complex64"},
					},
				},
			},
		},
		{
			name: "complex128",
			id:   &ast.Ident{Name: "c"},
			typ:  &ast.Ident{Name: "complex128"},
			expectedExpr: &ast.KeyValueExpr{
				Key: &ast.Ident{Name: "c"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "value"},
						Sel: &ast.Ident{Name: "Complex128"},
					},
				},
			},
		},
		{
			name: "Struct_SamePackage",
			id:   &ast.Ident{Name: "a"},
			typ:  &ast.Ident{Name: "Address"},
			expectedExpr: &ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "a"},
				Value: &ast.Ident{Name: "nil"},
			},
		},
		{
			name: "Struct_OtherPackage",
			id:   &ast.Ident{Name: "t"},
			typ: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "http"},
				Sel: &ast.Ident{Name: "Transport"},
			},
			expectedExpr: &ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "t"},
				Value: &ast.Ident{Name: "nil"},
			},
		},
		{
			name: "Pointer",
			id:   &ast.Ident{Name: "p"},
			typ: &ast.StarExpr{
				X: &ast.Ident{Name: "int"},
			},
			expectedExpr: &ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "p"},
				Value: &ast.Ident{Name: "nil"},
			},
		},
		{
			name: "Slice",
			id:   &ast.Ident{Name: "s"},
			typ: &ast.ArrayType{
				Elt: &ast.Ident{Name: "int"},
			},
			expectedExpr: &ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "s"},
				Value: &ast.Ident{Name: "nil"},
			},
		},
		{
			name: "Map",
			id:   &ast.Ident{Name: "m"},
			typ: &ast.MapType{
				Key:   &ast.Ident{Name: "int"},
				Value: &ast.Ident{Name: "string"},
			},
			expectedExpr: &ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "m"},
				Value: &ast.Ident{Name: "nil"},
			},
		},
		{
			name: "Channel",
			id:   &ast.Ident{Name: "c"},
			typ: &ast.ChanType{
				Value: &ast.Ident{Name: "error"},
			},
			expectedExpr: &ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "c"},
				Value: &ast.Ident{Name: "nil"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expr := createFieldInitExpr(tc.id, tc.typ)

			assert.Equal(t, tc.expectedExpr, expr)
		})
	}
}
