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
	prefixes := []string{"Test", "Fuzz", "Example"}
	fset := token.NewFileSet()
	var found []string
	paths := bufio.NewScanner(os.Stdin)
	for paths.Scan() {
		file, err := parser.ParseFile(fset, paths.Text(), nil, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && (fn.Recv != nil || !slices.ContainsFunc(prefixes, func(p string) bool { return strings.HasPrefix(fn.Name.Name, p) })) {
				found = append(found, fmt.Sprintf("%s: func %s", fset.Position(fn.Pos()), fn.Name.Name))
			}
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) > 0 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if ok && sel.Sel.Name == "Helper" {
				found = append(found, fmt.Sprintf("%s: Helper()", fset.Position(call.Pos())))
			}
			return true
		})
	}
	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "test helpers are prohibited; a test file declares only Test, Fuzz and Example functions (.shampoo/rules/test-helpers.toml):")
		fmt.Fprintln(os.Stderr, strings.Join(found, "\n"))
		os.Exit(1)
	}
}
