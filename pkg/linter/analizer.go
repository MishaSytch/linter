package linter

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"linter/pkg/linter/util"
)

// Analyzer определяет конфигурацию линтера для использования в основной программе.
var Analyzer = &analysis.Analyzer{
	Name: "loglint",
	Doc:  `Описание`, //FIXME: написать нужно описание
	Run:  run,
}

// ArgumentAnalyzer функциональный тип для анализа кода
type ArgumentAnalyzer func(pass *analysis.Pass, arg ast.Expr)

func run(pass *analysis.Pass) (interface{}, error) {
	var checkArg ArgumentAnalyzer = util.BaseAnalyzer
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 || !util.IsLogCall(call) {
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
