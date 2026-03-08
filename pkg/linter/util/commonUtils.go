package util

import (
	"github.com/MishaSytch/linter/pkg/linter/model"
	"go/ast"
	"strings"
)

var fieldConstructors = []string{"Object", "String", "Int", "Any", "AddString", "AddObject"}

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

	if selection, ok := pass.GetSelection(selector); ok {
		if recv := selection.Recv(); recv != nil {
			recvStr := recv.String()
			for _, name := range model.LoggerList {
				if strings.Contains(strings.ToLower(recvStr), strings.ToLower(name)) {
					return true
				}
			}
		}
	}

	if tv, ok := pass.GetTypeAndValue(selector.X); ok && tv.Type != nil {
		typeStr := tv.Type.String()
		for _, name := range model.LoggerList {
			if strings.Contains(strings.ToLower(typeStr), strings.ToLower(name)) {
				return true
			}
		}
	}

	root := getRootResolver(selector.X)

	if id, ok := root.(*ast.Ident); ok {
		for _, name := range model.LoggerList {
			if id.Name == name {
				return true
			}
		}
	}

	if tv, ok := pass.GetTypeAndValue(root); ok {
		typeStr := tv.Type.String()
		for _, name := range model.LoggerList {
			if strings.Contains(strings.ToLower(typeStr), strings.ToLower(name)) {
				return true
			}
		}
	}

	if tv, ok := pass.GetTypeAndValue(call.Fun); ok {
		typeStr := tv.Type.String()
		if strings.Contains(typeStr, "Field") {
			for _, name := range model.LoggerList {
				if strings.Contains(strings.ToLower(typeStr), strings.ToLower(name)) {
					return true
				}
			}
		}
	}

	for _, name := range fieldConstructors {
		if selector.Sel.Name == name {
			return true
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
