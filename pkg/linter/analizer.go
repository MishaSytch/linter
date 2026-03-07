package linter

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"linter/pkg/config"
	"linter/pkg/linter/util"
	"linter/pkg/linter/util/baseAnalizer"
)

// Analyzer определяет конфигурацию линтера для использования в основной программе.
var Analyzer = &analysis.Analyzer{
	Name: "loglint",
	Doc:  `Описание`, //FIXME: написать нужно описание
	Run:  run,
}

// ArgumentAnalyzer функциональный тип для анализа кода
type ArgumentAnalyzer func(pass *analysis.Pass, arg ast.Expr, config *config.Config)

var configPath string
var cfg *config.Config

func init() {
	Analyzer.Flags.StringVar(&configPath, "config", "config.yml", "path to config file")
}

func run(pass *analysis.Pass) (interface{}, error) {
	cfg = config.LoadConfig(configPath)

	var checkArg ArgumentAnalyzer = baseAnalizer.BaseAnalyzer
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 || !util.IsLogCall(call) {
				return true
			}

			for _, arg := range call.Args {
				checkArg(pass, arg, cfg)
			}
			return true
		})
	}
	return nil, nil
}
