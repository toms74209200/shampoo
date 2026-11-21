package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	var found []string
	paths := bufio.NewScanner(os.Stdin)
	for paths.Scan() {
		file, err := parser.ParseFile(fset, paths.Text(), nil, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, lit := range Literals(file) {
			found = append(found, fmt.Sprintf("%s: %s", fset.Position(lit.Pos()), lit.Value))
		}
	}
	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "literal values in tests appear only in the cases of a parameterized test table (.shampoo/rules/fixed-values.toml):")
		fmt.Fprintln(os.Stderr, strings.Join(found, "\n"))
		os.Exit(1)
	}
}

func Literals(file *ast.File) []*ast.BasicLit {
	messages := []string{"Error", "Errorf", "Fatal", "Fatalf", "Log", "Logf", "Skip", "Skipf"}
	var found []*ast.BasicLit
	var stack []ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		lit, isLit := n.(*ast.BasicLit)
		allowed := slices.ContainsFunc(stack, func(parent ast.Node) bool {
			switch p := parent.(type) {
			case *ast.ImportSpec:
				return true
			case *ast.CompositeLit:
				switch p.Type.(type) {
				case *ast.ArrayType, *ast.MapType:
					return true
				}
			case *ast.CallExpr:
				sel, ok := p.Fun.(*ast.SelectorExpr)
				return ok && (slices.Contains(messages, sel.Sel.Name) || sel.Sel.Name == "Run" && len(p.Args) == 2 && p.Args[0] == ast.Node(lit))
			}
			return false
		})
		if isLit && !allowed {
			found = append(found, lit)
		}
		stack = append(stack, n)
		return true
	})
	return found
}
