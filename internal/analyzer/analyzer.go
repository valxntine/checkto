package analyzer

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "gochecktimeout",
	Doc:      "Checks struct timeout fields use time.Duration and timeout assignments don't use operations",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.AssignStmt)(nil),
		(*ast.TypeSpec)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		checkFields(node, pass)
		checkAssignment(node, pass)
		return
	})

	return nil, nil
}

func checkAssignment(node ast.Node, pass *analysis.Pass) {
	assignStmt, ok := node.(*ast.AssignStmt)
	if !ok {
		return
	}

	compLit, ok := assignStmt.Rhs[0].(*ast.CompositeLit)
	if !ok {
		return
	}

	if compLit.Elts == nil {
		return
	}

	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		k := kv.Key.(*ast.Ident).Name

		if strings.Contains(strings.ToLower(k), "timeout") {
			val, ok := kv.Value.(*ast.BinaryExpr)
			if !ok {
				continue
			}

			var firstParam string
			switch val.X.(type) {
			case *ast.Ident:
				firstParam = val.X.(*ast.Ident).Name
			case *ast.SelectorExpr:
				firstParam = fmt.Sprintf("%s.%s", val.X.(*ast.SelectorExpr).X.(*ast.Ident).Name, val.X.(*ast.SelectorExpr).Sel.Name)
			}

			var secondParam string
			switch val.Y.(type) {
			case *ast.Ident:
				secondParam = val.Y.(*ast.Ident).Name
			case *ast.SelectorExpr:
				secondParam = fmt.Sprintf("%s.%s", val.Y.(*ast.SelectorExpr).X.(*ast.Ident).Name, val.Y.(*ast.SelectorExpr).Sel.Name)
			}

			pass.Reportf(node.Pos(), "assignment to %s contains operation %s %s %s but should use defined time.Duration",
				k,
				firstParam,
				val.Op.String(),
				secondParam,
			)
		}
	}
	return

}

func checkFields(node ast.Node, pass *analysis.Pass) {
	typeSpec, ok := node.(*ast.TypeSpec)
	if !ok {
		return
	}
	structDef, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return
	}
	fields := structDef.Fields.List

	for _, f := range fields {
		if len(f.Names) > 0 && strings.Contains(strings.ToLower(f.Names[0].Name), "timeout") {
			selectorExpr, ok := f.Type.(*ast.SelectorExpr)
			if !ok {
				ident, ok := f.Type.(*ast.Ident)
				if !ok {
					continue
				}

				pass.Reportf(
					f.Pos(),
					"timeout field %s should use time.Duration instead of %s",
					f.Names[0].Name,
					ident.Name,
				)
				continue
			}

			x, ok := selectorExpr.X.(*ast.Ident)
			if !ok || x.Name != "time" || selectorExpr.Sel.Name != "Duration" {
				pass.Reportf(
					f.Pos(),
					"timeout field %s should use time.Duration",
					f.Names[0].Name,
				)
			}
		}
	}

	return
}
