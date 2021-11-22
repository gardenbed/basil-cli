package builder

import "go/ast"

func createFieldInitExpr(id *ast.Ident, typ ast.Expr) *ast.KeyValueExpr {
	var value ast.Expr

	switch e := typ.(type) {
	case *ast.Ident:
		switch e.Name {
		case "error":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Error"},
				},
			}

		case "bool":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Bool"},
				},
			}

		case "string":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "String"},
				},
			}

		case "byte":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Byte"},
				},
			}

		case "rune":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Rune"},
				},
			}

		case "int":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Int"},
				},
			}

		case "int8":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Int8"},
				},
			}

		case "int16":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Int16"},
				},
			}

		case "int32":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Int32"},
				},
			}

		case "int64":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Int64"},
				},
			}

		case "uint":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Uint"},
				},
			}

		case "uint8":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Uint8"},
				},
			}

		case "uint16":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Uint16"},
				},
			}

		case "uint32":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Uint32"},
				},
			}

		case "uint64":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Uint64"},
				},
			}

		case "uintptr":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Uintptr"},
				},
			}

		case "float32":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Float32"},
				},
			}

		case "float64":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Float64"},
				},
			}

		case "complex64":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Complex64"},
				},
			}

		case "complex128":
			value = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "value"},
					Sel: &ast.Ident{Name: "Complex128"},
				},
			}

		// struct
		default:
			// TODO:
			value = &ast.Ident{Name: "nil"}
		}

	case *ast.SelectorExpr:
		// TODO:
		value = &ast.Ident{Name: "nil"}

	case *ast.StarExpr:
		// TODO:
		value = &ast.Ident{Name: "nil"}

	case *ast.ArrayType:
		// TODO:
		value = &ast.Ident{Name: "nil"}

	case *ast.MapType:
		// TODO:
		value = &ast.Ident{Name: "nil"}

	case *ast.ChanType:
		// TODO:
		value = &ast.Ident{Name: "nil"}
	}

	return &ast.KeyValueExpr{
		Key:   &ast.Ident{Name: id.Name},
		Value: value,
	}
}
