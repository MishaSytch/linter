package baseAnalizer

import (
	"github.com/MishaSytch/linter/pkg/linter/model"
	"go/ast"
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
