package util

import (
	"github.com/MishaSytch/linter/pkg/linter/model"
	"go/ast"
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

// IsLogCall определяет, является ли вызов функции обращением к одному из
// поддерживаемых логгеров, указанных в model.LoggerList
func IsLogCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	for _, loggerName := range model.LoggerList {
		if x.Name == loggerName {
			return true
		}
	}

	return false
}
