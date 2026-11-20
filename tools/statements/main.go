package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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
			os.Exit(2)
		}
		for _, branch := range Branches(file) {
			found = append(found, fmt.Sprintf("%s: %s", fset.Position(branch.Pos), branch.Kind))
		}
	}
	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "side-effect code is statements only; a choice other than handling an error or matching a sum type belongs in pure code (.shampoo/rules/side-effect-statements.toml):")
		fmt.Fprintln(os.Stderr, strings.Join(found, "\n"))
		os.Exit(1)
	}
}

type Branch struct {
	Pos  token.Pos
	Kind string
}

func Branches(file *ast.File) []Branch {
	var errorOnly func(ast.Expr) bool
	errorOnly = func(expr ast.Expr) bool {
		switch e := expr.(type) {
		case *ast.ParenExpr:
			return errorOnly(e.X)
		case *ast.UnaryExpr:
			return e.Op == token.NOT && errorOnly(e.X)
		case *ast.BinaryExpr:
			if e.Op == token.LAND || e.Op == token.LOR {
				return errorOnly(e.X) && errorOnly(e.Y)
			}
			x, xIdent := e.X.(*ast.Ident)
			y, yIdent := e.Y.(*ast.Ident)
			return (e.Op == token.NEQ || e.Op == token.EQL) && xIdent && yIdent && x.Name == "err" && y.Name == "nil"
		case *ast.CallExpr:
			fun, selected := e.Fun.(*ast.SelectorExpr)
			pkg, named := ast.Expr(nil), false
			if selected {
				pkg = fun.X
				_, named = pkg.(*ast.Ident)
			}
			return named && pkg.(*ast.Ident).Name == "errors" && (fun.Sel.Name == "Is" || fun.Sel.Name == "As")
		}
		return false
	}
	var found []Branch
	ast.Inspect(file, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.IfStmt:
			if !errorOnly(s.Cond) {
				found = append(found, Branch{Pos: s.Pos(), Kind: "if"})
			}
		case *ast.SwitchStmt:
			found = append(found, Branch{Pos: s.Pos(), Kind: "switch"})
		case *ast.SelectStmt:
			found = append(found, Branch{Pos: s.Pos(), Kind: "select"})
		case *ast.ForStmt:
			found = append(found, Branch{Pos: s.Pos(), Kind: "for"})
		}
		return true
	})
	return found
}
