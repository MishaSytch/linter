package baseAnalizer

import (
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
)

// Reporter интерфейс заменяет стандартный *analysis.Pass и используется
// благодаря утиной типизации
type Reporter interface {
	// Reportf регистрирует предупреждение в месте pos с форматированным сообщением
	Reportf(pos token.Pos, format string, args ...interface{})

	// TypesInfo возвращает информацию о типах узлов
	TypesInfo() *types.Info
}

// PassWrapper обертка над analysis.Pass
type PassWrapper struct {
	*analysis.Pass
}

func (w PassWrapper) TypesInfo() *types.Info {
	return w.Pass.TypesInfo
}

type errorsChecker func(pass Reporter, arg ast.Expr, msg string)
type textErrors map[string]bool
type textSyntaxChecker func(pass Reporter, arg ast.Expr, errors textErrors, r rune)
