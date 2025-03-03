package analyzer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/valxntine/checkto/internal/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %s", err)
	}

	testdata := filepath.Join(filepath.Dir(filepath.Dir(wd)), "testdata")
	analysistest.Run(t, testdata, analyzer.Analyzer, "t")
}
