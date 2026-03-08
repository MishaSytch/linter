package model

import (
	"go/ast"
)

// FunctionalAnalyzer интерфейс для анализаторов кода
type FunctionalAnalyzer interface {
	// Analyze запуск анализа
	Analyze(arg ast.Expr)
}
