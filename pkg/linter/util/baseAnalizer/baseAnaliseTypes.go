package baseAnalizer

import (
	"go/ast"
)

type analysisResult struct {
	originalMsg string
	fixedMsg    string
	errors      []string
}

type errorsCheckerAggregate func(res *analysisResult)
type errorsChecker func(arg ast.Expr, msg string)
type textErrors map[string]bool
type textSyntaxChecker func(arg ast.Expr, errors textErrors, r rune)
type textSyntaxCheckerWithBool func(res *analysisResult, r rune, errors textErrors) bool
