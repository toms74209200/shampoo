package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	pattern := regexp.MustCompile(`^when \S.* with \S.* then returns \S.*$`)
	fset := token.NewFileSet()
	var found []string
	paths := bufio.NewScanner(os.Stdin)
	for paths.Scan() {
		file, err := parser.ParseFile(fset, paths.Text(), nil, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fields := map[string]bool{}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(fn.Name.Name, "Test") || fn.Name.Name == "TestMain" {
				continue
			}
			runs := 0
			ast.Inspect(fn, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) != 2 {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Run" {
					return true
				}
				runs++
				switch name := call.Args[0].(type) {
				case *ast.BasicLit:
					text, _ := strconv.Unquote(name.Value)
					if !pattern.MatchString(text) {
						found = append(found, fmt.Sprintf("%s: %s", fset.Position(name.Pos()), name.Value))
					}
				case *ast.SelectorExpr:
					fields[name.Sel.Name] = true
				default:
					found = append(found, fmt.Sprintf("%s: a test name is a string literal or a table field", fset.Position(name.Pos())))
				}
				return true
			})
			if runs == 0 {
				found = append(found, fmt.Sprintf("%s: %s names no test with t.Run", fset.Position(fn.Pos()), fn.Name.Name))
			}
		}
		ast.Inspect(file, func(n ast.Node) bool {
			kv, ok := n.(*ast.KeyValueExpr)
			if !ok {
				return true
			}
			key, ok := kv.Key.(*ast.Ident)
			lit, isLit := kv.Value.(*ast.BasicLit)
			if ok && fields[key.Name] && isLit {
				text, _ := strconv.Unquote(lit.Value)
				if !pattern.MatchString(text) {
					found = append(found, fmt.Sprintf("%s: %s", fset.Position(lit.Pos()), lit.Value))
				}
			}
			return true
		})
	}
	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "every test is named `when <target> with <given conditions> then returns <expected>` (.shampoo/rules/test-names.toml):")
		fmt.Fprintln(os.Stderr, strings.Join(found, "\n"))
		os.Exit(1)
	}
}
