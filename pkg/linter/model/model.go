package model

import (
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

var _ Reporter = (*PassWrapper)(nil)
