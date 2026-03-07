package baseAnalizer

import (
	"go/ast"
	"linter/pkg/linter/model"
)

type analysisResult struct {
	originalMsg string
	fixedMsg    string
	errors      []string
}

type errorsCheckerAggregate func(res *analysisResult)
type errorsChecker func(pass model.Reporter, arg ast.Expr, msg string)
type textErrors map[string]bool
type textSyntaxChecker func(pass model.Reporter, arg ast.Expr, errors textErrors, r rune)
type textSyntaxCheckerWithBool func(res *analysisResult, r rune, errors textErrors) bool
