package linter

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"log/slog"
)

// Analyzer определяет конфигурацию линтера для использования в основной программе.
var Analyzer = &analysis.Analyzer{
	Name: "loglint",
	Doc:  `Описание`, //FIXME: написать нужно описание
	Run:  run,
}

var checkArg func(arg ast.Expr)

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}

			checkArg = func(arg ast.Expr) {
				slog.Error("Оно не работает пока!")
			}

			for _, arg := range call.Args {
				checkArg(arg)
			}
			return true
		})
	}
	return nil, nil
}
