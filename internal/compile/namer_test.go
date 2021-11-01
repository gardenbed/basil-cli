package compile

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsExported(t *testing.T) {
	tests := []struct {
		name           string
		expectedResult bool
	}{
		{"internal", false},
		{"External", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsExported(tc.name)

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestConvertToUnexported(t *testing.T) {
	tests := []struct {
		name         string
		expectedName string
	}{
		{
			name:         "err",
			expectedName: "err",
		},
		{
			name:         "ID",
			expectedName: "id",
		},
		{
			name:         "URL",
			expectedName: "url",
		},
		{
			name:         "User",
			expectedName: "user",
		},
		{
			name:         "UserID",
			expectedName: "userID",
		},
		{
			name:         "HTTPRequest",
			expectedName: "httpRequest",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name := ConvertToUnexported(tc.name)

			assert.Equal(t, tc.expectedName, name)
		})
	}
}

func TestInferName(t *testing.T) {
	tests := []struct {
		name        string
		expr        ast.Expr
		expecteName string
	}{
		{
			name:        "ExportedStruct",
			expr:        &ast.Ident{Name: "Embedded"},
			expecteName: "Embedded",
		},
		{
			name:        "UnexportedStruct",
			expr:        &ast.Ident{Name: "embedded"},
			expecteName: "embedded",
		},
		{
			name: "PackageStruct",
			expr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "entity"},
				Sel: &ast.Ident{Name: "Embedded"},
			},
			expecteName: "Embedded",
		},
		{
			name: "PackageStruct",
			expr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "entity"},
				Sel: &ast.Ident{Name: "Embedded"},
			},
			expecteName: "Embedded",
		},
		{
			name: "PointerPackageStruct",
			expr: &ast.StarExpr{
				X: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "entity"},
					Sel: &ast.Ident{Name: "Embedded"},
				},
			},
			expecteName: "Embedded",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name := InferName(tc.expr)

			assert.Equal(t, tc.expecteName, name)
		})
	}
}
