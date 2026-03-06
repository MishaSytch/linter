package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"
	"linter/pkg/linter"
	"log/slog"
)

func main() {
	singlechecker.Main(linter.Analyzer)
	slog.Info(linter.Analyzer.Doc)
}
