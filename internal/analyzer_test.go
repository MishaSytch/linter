package internal

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"linter/pkg/linter"
	"testing"
)

func TestAll(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, linter.Analyzer, "test_case")
}
