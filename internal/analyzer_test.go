package internal

import (
	"github.com/MishaSytch/linter/pkg/linter"
	"golang.org/x/tools/go/analysis/analysistest"
	"os"
	"path/filepath"
	"testing"
)

func TestAll(t *testing.T) {
	wd, _ := os.Getwd()
	testdata := filepath.Join(wd, "testdata")
	analysistest.Run(t, testdata, linter.Analyzer, "test_case")
}
