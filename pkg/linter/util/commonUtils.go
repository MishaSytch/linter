package util

import (
	"github.com/MishaSytch/linter/pkg/linter/model"
	"go/ast"
	"strings"
)

// IsFmtSprintf проверяет, является ли выражение вызовом функции fmt.Sprintf()
func IsFmtSprintf(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := selector.X.(*ast.Ident)
	return ok && x.Name == "fmt" && selector.Sel.Name == "Sprintf"
}

// IsLogCall определяет, является ли вызов функции обращением к логгеру.
func IsLogCall(pass model.Reporter, call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if x, ok := selector.X.(*ast.Ident); ok {
		for _, name := range model.LoggerList {
			if x.Name == name {
				return true
			}
		}
	}

	if selection, ok := pass.GetSelection(selector); ok {
		recv := selection.Recv()
		if recv != nil {
			recvStr := recv.String()
			for _, name := range model.LoggerList {
				if strings.Contains(recvStr, name) {
					return true
				}
			}
		}
	}

	if tv, ok := pass.GetTypeAndValue(call.Fun); ok {
		typeStr := tv.Type.String()
		for _, name := range model.LoggerList {
			if strings.Contains(typeStr, name) && strings.Contains(typeStr, "Field") {
				return true
			}
		}
	}

	if obj, ok := pass.GetObject(selector.Sel); ok {
		if pkg := obj.Pkg(); pkg != nil {
			pkgPath := pkg.Path()
			for _, name := range model.LoggerList {
				if strings.Contains(pkgPath, name) {
					return true
				}
			}
		}
	}

	return false
}

// getRootResolver рекурсивно находит корень цепочки вызовов
func getRootResolver(expr ast.Expr) ast.Expr {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		return getRootResolver(t.X)
	case *ast.CallExpr:
		return getRootResolver(t.Fun)
	default:
		return t
	}
}
