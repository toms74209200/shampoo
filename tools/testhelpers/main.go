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
		for _, helper := range Helpers(file) {
			found = append(found, fmt.Sprintf("%s: %s", fset.Position(helper.Pos), helper.Name))
		}
	}
	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "test helpers are prohibited; a test file declares only Test, Fuzz and Example functions (.shampoo/rules/test-helpers.toml):")
		fmt.Fprintln(os.Stderr, strings.Join(found, "\n"))
		os.Exit(1)
	}
}

type Helper struct {
	Pos  token.Pos
	Name string
}

func Helpers(file *ast.File) []Helper {
	prefixes := []string{"Test", "Fuzz", "Example"}
	var found []Helper
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && (fn.Recv != nil || !slices.ContainsFunc(prefixes, func(p string) bool { return strings.HasPrefix(fn.Name.Name, p) })) {
			found = append(found, Helper{Pos: fn.Pos(), Name: "func " + fn.Name.Name})
		}
	}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) > 0 {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok && sel.Sel.Name == "Helper" {
			found = append(found, Helper{Pos: call.Pos(), Name: "Helper()"})
		}
		return true
	})
	return found
}
