package internal

import (
	"github.com/MishaSytch/linter/pkg/linter"
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestAll(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, linter.Analyzer, "test_case")
}
