package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	name := regexp.MustCompile(`^when (\S+) with `)
	fset := token.NewFileSet()
	production := map[string]map[string]bool{}
	tests := map[string][]*ast.File{}
	paths := bufio.NewScanner(os.Stdin)
	for paths.Scan() {
		file, err := parser.ParseFile(fset, paths.Text(), nil, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		dir := filepath.Dir(paths.Text())
		if strings.HasSuffix(paths.Text(), "_test.go") {
			if !strings.HasSuffix(file.Name.Name, "_test") {
				tests[dir] = append(tests[dir], file)
			}
			continue
		}
		if production[dir] == nil {
			production[dir] = map[string]bool{}
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					production[dir][d.Name.Name] = true
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if v, ok := spec.(*ast.ValueSpec); ok {
						for _, id := range v.Names {
							production[dir][id.Name] = true
						}
					}
				}
			}
		}
	}
	var found []string
	for dir, files := range tests {
		for _, file := range files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				targets := map[string]bool{}
				ast.Inspect(fn, func(n ast.Node) bool {
					if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
						text, _ := strconv.Unquote(lit.Value)
						if m := name.FindStringSubmatch(text); m != nil {
							targets[m[1]] = true
						}
					}
					return true
				})
				names := map[*ast.Ident]bool{}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					switch node := n.(type) {
					case *ast.SelectorExpr:
						names[node.Sel] = true
					case *ast.KeyValueExpr:
						if key, ok := node.Key.(*ast.Ident); ok {
							names[key] = true
						}
					}
					return true
				})
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					id, ok := n.(*ast.Ident)
					if ok && !names[id] && production[dir][id.Name] && !targets[id.Name] {
						found = append(found, fmt.Sprintf("%s: %s uses %s", fset.Position(id.Pos()), fn.Name.Name, id.Name))
					}
					return true
				})
			}
		}
	}
	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "a white-box test depends on its target only, and on types for its arguments (.shampoo/rules/white-box-dependencies.toml):")
		fmt.Fprintln(os.Stderr, strings.Join(found, "\n"))
		os.Exit(1)
	}
}
