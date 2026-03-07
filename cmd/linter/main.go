package main

import (
	"github.com/MishaSytch/linter/pkg/linter"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(linter.Analyzer)
}
