package main

import (
	checkto "github.com/valxntine/checkto"
	"golang.org/x/tools/go/analysis"
)

func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{checkto.DurationAnalyzer}, nil
}
