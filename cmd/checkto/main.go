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

	var keys []string
	var ops []string
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

			sel, ok := val.Y.(*ast.SelectorExpr)
			if !ok {
				 continue
			}

			keys = append(keys, k)

			s := fmt.Sprintf("%s %s %s.%s",
				val.X.(*ast.Ident).Name,
				val.Op.String(),
				sel.X.(*ast.Ident).Name,
				sel.Sel.Name,
			)
			ops = append(ops, s)
		}
	}


	for i := range ops {
		fmt.Printf(
			"%s: assignment to %s contains operation '%s' - should use defined time.Duration\n",
			v.fset.Position(node.Pos()),
			keys[i],
			ops[i],
		)
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
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f.Names[0].Name), "timeout"){
			timeoutFields = append(timeoutFields, f.Names[0].Name)
		}
	}
	if len(timeoutFields) == 0 {
		return
	}

	pre := fmt.Sprintf("%s: struct '%s' contains timeout field", v.fset.Position(node.Pos()), typeSpec.Name.Name)
	jointFields := strings.Join(timeoutFields, ", ")

	var sb strings.Builder
	sb.WriteString(pre)
	if len(timeoutFields) > 1 {
		sb.WriteString("s")
	}
	sb.WriteString(" [")
	sb.WriteString(jointFields)
	sb.WriteString("] which do")
	if len(timeoutFields) == 1 {
		sb.WriteString("es")
	}
	sb.WriteString(" not use time.Duration as the type")
	fmt.Println(sb.String())
	return
}
