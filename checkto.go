package checkto

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var DurationAnalyzer = &analysis.Analyzer{
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

	// Check if Rhs has at least one element
	if len(assignStmt.Rhs) == 0 {
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

		// Safely extract the key name
		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok {
			// Key is not a simple identifier (could be selector, etc.)
			continue
		}
		k := keyIdent.Name

		if strings.Contains(strings.ToLower(k), "timeout") {
			// Unwrap parenthesized expressions
			expr := kv.Value
			for {
				if parenExpr, ok := expr.(*ast.ParenExpr); ok {
					expr = parenExpr.X
				} else {
					break
				}
			}

			val, ok := expr.(*ast.BinaryExpr)
			if !ok {
				continue
			}

			firstParam := exprToString(val.X)
			secondParam := exprToString(val.Y)

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

// exprToString converts an expression to a string representation
// Handles various expression types safely without panicking
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		// Recursively handle nested selectors like cfg.Inner.Timeout
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.BasicLit:
		// Integer/float literals like 30, 2.5
		return e.Value
	case *ast.CallExpr:
		// Function calls like getTimeout() or time.Duration(x)
		return "<function call>"
	case *ast.ParenExpr:
		// Parenthesized expressions - unwrap them
		return exprToString(e.X)
	default:
		// Other expression types (unary, type asserts, etc.)
		return "<expression>"
	}
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
		// Handle multiple field names on one line (e.g., Start, End, Timeout int)
		for _, fieldName := range f.Names {
			if !strings.Contains(strings.ToLower(fieldName.Name), "timeout") {
				continue
			}

			// Check if it's time.Duration
			if isTimeDuration(f.Type) {
				// Correct type, no warning needed
				continue
			}

			// It's a timeout field but not time.Duration - report it
			typeName := typeToString(f.Type)
			if typeName != "" {
				pass.Reportf(
					f.Pos(),
					"timeout field %s should use time.Duration instead of %s",
					fieldName.Name,
					typeName,
				)
			} else {
				pass.Reportf(
					f.Pos(),
					"timeout field %s should use time.Duration",
					fieldName.Name,
				)
			}
		}
	}

	return
}

// isTimeDuration checks if a type expression is time.Duration or *time.Duration
func isTimeDuration(expr ast.Expr) bool {
	// Handle pointer types (*time.Duration)
	if starExpr, ok := expr.(*ast.StarExpr); ok {
		expr = starExpr.X
	}

	selectorExpr, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	x, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	return x.Name == "time" && selectorExpr.Sel.Name == "Duration"
}

// typeToString converts a type expression to a readable string
func typeToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		if x, ok := e.X.(*ast.Ident); ok {
			return x.Name + "." + e.Sel.Name
		}
		return e.Sel.Name
	case *ast.StarExpr:
		return "*" + typeToString(e.X)
	case *ast.ArrayType:
		return "[]" + typeToString(e.Elt)
	default:
		return ""
	}
}
