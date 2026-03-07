package linter

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// Analyzer определяет конфигурацию линтера для использования в основной программе.
var Analyzer = &analysis.Analyzer{
	Name: "loglint",
	Doc:  `Описание`, //FIXME: написать нужно описание
	Run:  run,
}

type ArgumentAnalyzer func(pass *analysis.Pass, arg ast.Expr)

func run(pass *analysis.Pass) (interface{}, error) {
	var checkArg ArgumentAnalyzer = BaseAnalyzer

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}

			for _, arg := range call.Args {
				checkArg(pass, arg)
			}
			return true
		})
	}
	return nil, nil
}
