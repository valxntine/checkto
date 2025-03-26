package checkto_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valxntine/checkto"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestDurationAnalyzer(t *testing.T) {
	analysistest.Run(t, testDataDir(t), checkto.DurationAnalyzer, "t")
}

func testDataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(1)
	if !ok {
		require.Fail(t, "unable to get current test file name")
	}

	return filepath.Join(filepath.Dir(testFilename), "testdata")
}
