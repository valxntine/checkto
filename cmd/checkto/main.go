package main

import (
	"github.com/valxntine/checkto/internal/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main () {
	singlechecker.Main(analyzer.Analyzer)
}
