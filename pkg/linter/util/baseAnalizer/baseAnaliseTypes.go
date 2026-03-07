package baseAnalizer

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

type errorsChecker func(pass *analysis.Pass, arg ast.Expr, msg string)
type textErrors map[string]bool
type textSyntaxChecker func(pass *analysis.Pass, arg ast.Expr, errors textErrors, r rune)
