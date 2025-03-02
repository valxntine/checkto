package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

type visitor struct {
	fset *token.FileSet
}

func main() {
	v := visitor{fset: token.NewFileSet()}
	for _, filePath := range os.Args[1:] {
		if filePath == "--" {
			continue
		}

		f, err := parser.ParseFile(v.fset, filePath, nil, 0)
		if err != nil {
			log.Fatalf("failed to parse: %s - %s", filePath, err)
		}

		ast.Walk(&v, f)
	}
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	v.checkFields(node)
	v.checkAssignment(node)
	return v
}

func (v *visitor) checkAssignment(node ast.Node) {
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

			fmt.Printf("%s: assignment to %s contains operation '%s %s %s' - should use defined time.Duration\n",
				v.fset.Position(node.Pos()),
				k,
				firstParam,
				val.Op.String(),
				secondParam,
			)
		}
	}
	return

}

func (v *visitor) checkFields(node ast.Node) {
	typeSpec, ok := node.(*ast.TypeSpec)
	if !ok {
		return
	}
	structDef, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return
	}
	fields := structDef.Fields.List
	var timeoutFields []string
	var timeoutFieldVals []string
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f.Names[0].Name), "timeout") {
			fType, ok := f.Type.(*ast.SelectorExpr)
			if !ok {
				timeoutFields = append(timeoutFields, f.Names[0].Name)
				timeoutFieldVals = append(timeoutFieldVals, f.Type.(*ast.Ident).Name)
				continue
			}
			if fType.X.(*ast.Ident).Name != "time" || fType.Sel.Name != "Duration" {
				timeoutFields = append(timeoutFields, f.Names[0].Name)
				timeoutFieldVals = append(timeoutFieldVals, fmt.Sprintf("%s.%s", fType.X.(*ast.Ident).Name, fType.Sel.Name))
			}
		}
	}
	if len(timeoutFields) == 0 {
		return
	}

	for i := range timeoutFields {
		fmt.Printf(
			"%s: struct '%s' contains timeout field '%s' which does not use time.Duration as the type (uses %s)\n",
			v.fset.Position(node.Pos()),
			typeSpec.Name.Name,
			timeoutFields[i],
			timeoutFieldVals[i],
		)
	}

	return
}
