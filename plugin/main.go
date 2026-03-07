package main

import (
	"github.com/MishaSytch/linter/pkg/linter"
	"golang.org/x/tools/go/analysis"
)

type analyzerPlugin struct{}

func (p *analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		linter.Analyzer,
	}
}

func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{linter.Analyzer}, nil
}

var AnalyzerPlugin analyzerPlugin

func main() {}
