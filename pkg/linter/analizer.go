package linter

import (
	"github.com/MishaSytch/linter/pkg/config"
	"github.com/MishaSytch/linter/pkg/linter/model"
	"github.com/MishaSytch/linter/pkg/linter/util"
	"github.com/MishaSytch/linter/pkg/linter/util/baseAnalizer"
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// Analyzer определяет конфигурацию линтера для использования в основной программе.
var Analyzer = &analysis.Analyzer{
	Name: "loglint",
	Doc:  `Описание`, //FIXME: написать нужно описание
	Run:  run,
}

// ArgumentAnalyzer функциональный тип для анализа кода
type ArgumentAnalyzer func(pass model.Reporter, arg ast.Expr, config *config.Config)

var configPath string

func init() {
	Analyzer.Flags.StringVar(&configPath, "config", "config.yml", "path to config file")
}

func run(pass *analysis.Pass) (interface{}, error) {
	cfg := config.LoadConfig(configPath)

	wrapper := model.PassWrapper{Pass: pass}

	var checkArg ArgumentAnalyzer
	if cfg.Output.ErrorsAggregate {
		checkArg = baseAnalizer.BaseAggregateAnalyzer

	} else {
		checkArg = baseAnalizer.BaseAnalyzer
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 || !util.IsLogCall(call) {
				return true
			}

			for _, arg := range call.Args {
				checkArg(wrapper, arg, cfg)
			}
			return true
		})
	}
	return nil, nil
}
