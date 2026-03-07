package linter

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

func BaseAnalyzer(pass *analysis.Pass, arg ast.Expr) {
}
