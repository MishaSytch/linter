package model

import (
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
)

// SensitiveKeywords Чувствительные слова
var SensitiveKeywords = []string{
	"password",
	"token",
	"api_key",
	"secret",
}

// LoggerList Список анализируемых логеров
var LoggerList = []string{
	"log",
	"slog",
	"zap",
	"logger",
}

// Reporter интерфейс заменяет стандартный *analysis.Pass и используется
// благодаря утиной типизации
type Reporter interface {
	// Reportf регистрирует предупреждение в месте pos с форматированным сообщением
	Reportf(pos token.Pos, format string, args ...interface{})

	// Report регистрирует предупреждение в месте pos
	Report(d analysis.Diagnostic)

	// TypesInfo возвращает информацию о типах узлов
	TypesInfo() *types.Info
	ObjectOf(id *ast.Ident) types.Object
	TypeOf(expr ast.Expr) types.Type
	GetSelection(sel *ast.SelectorExpr) (*types.Selection, bool)
	GetTypeAndValue(expr ast.Expr) (types.TypeAndValue, bool)
	GetObject(id *ast.Ident) (types.Object, bool)
}

// PassWrapper обертка над analysis.Pass
type PassWrapper struct {
	*analysis.Pass
}

func (w PassWrapper) Report(d analysis.Diagnostic) {
	w.Pass.Report(d)
}

func (w PassWrapper) TypesInfo() *types.Info {
	return w.Pass.TypesInfo
}

func (w PassWrapper) ObjectOf(id *ast.Ident) types.Object {
	return w.Pass.TypesInfo.ObjectOf(id)
}

func (w PassWrapper) TypeOf(expr ast.Expr) types.Type {
	return w.Pass.TypesInfo.TypeOf(expr)
}

func (w PassWrapper) GetSelection(sel *ast.SelectorExpr) (*types.Selection, bool) {
	s, ok := w.Pass.TypesInfo.Selections[sel]
	return s, ok
}

func (w PassWrapper) GetTypeAndValue(expr ast.Expr) (types.TypeAndValue, bool) {
	tv, ok := w.Pass.TypesInfo.Types[expr]
	return tv, ok
}

func (w PassWrapper) GetObject(id *ast.Ident) (types.Object, bool) {
	obj, ok := w.Pass.TypesInfo.Uses[id]
	return obj, ok
}

var _ Reporter = (*PassWrapper)(nil)
