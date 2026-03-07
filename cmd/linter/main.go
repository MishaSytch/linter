package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"
	"linter/pkg/linter"
)

func main() {
	singlechecker.Main(linter.Analyzer)
}
