package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	var sources []Source
	paths := bufio.NewScanner(os.Stdin)
	for paths.Scan() {
		file, err := parser.ParseFile(fset, paths.Text(), nil, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		sources = append(sources, Source{Path: paths.Text(), File: file})
	}
	var found []string
	for _, finding := range Prohibited(sources) {
		found = append(found, fmt.Sprintf("%s: %s", fset.Position(finding.Pos), finding.Text))
	}
	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "prohibited functions; write them inline where they are used (.shampoo/rules/prohibited-functions.toml):")
		fmt.Fprintln(os.Stderr, strings.Join(found, "\n"))
		os.Exit(1)
	}
}

type Source struct {
	Path string
	File *ast.File
}

type Finding struct {
	Pos  token.Pos
	Text string
}

func Prohibited(sources []Source) []Finding {
	primitive := map[string]bool{"string": true, "bool": true, "byte": true, "rune": true, "error": true, "any": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
		"float32": true, "float64": true, "complex64": true, "complex128": true}
	packages := map[string][]*ast.File{}
	for _, source := range sources {
		dir := filepath.Dir(source.Path)
		packages[dir] = append(packages[dir], source.File)
	}
	var found []Finding
	for _, files := range packages {
		calls := map[string]int{}
		for _, file := range files {
			ast.Inspect(file, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					switch fun := call.Fun.(type) {
					case *ast.Ident:
						calls[fun.Name]++
					case *ast.SelectorExpr:
						calls[fun.Sel.Name]++
					}
				}
				return true
			})
		}
		for _, file := range files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil || fn.Name.Name == "main" || fn.Name.Name == "init" {
					continue
				}
				var reasons []string
				var params, results []string
				onlyPrimitive := fn.Type.Params.NumFields() > 0 && fn.Type.Results.NumFields() > 0
				for _, list := range []*ast.FieldList{fn.Type.Params, fn.Type.Results} {
					if list == nil {
						continue
					}
					for _, field := range list.List {
						var name strings.Builder
						ast.Inspect(field.Type, func(n ast.Node) bool {
							if id, ok := n.(*ast.Ident); ok {
								name.WriteString(id.Name)
							}
							return true
						})
						text := name.String()
						id, isIdent := field.Type.(*ast.Ident)
						onlyPrimitive = onlyPrimitive && isIdent && primitive[id.Name]
						for range max(len(field.Names), 1) {
							if list == fn.Type.Params {
								params = append(params, text)
							} else {
								results = append(results, text)
							}
						}
					}
				}
				if onlyPrimitive {
					reasons = append(reasons, "converts primitives into primitives")
				}
				for _, r := range results {
					for _, p := range params {
						if r == p && !primitive[r] {
							reasons = append(reasons, "returns the type it was given: "+r)
						}
					}
				}
				branches := 0
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					switch n.(type) {
					case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause, *ast.CommClause:
						branches++
					case *ast.FuncLit:
						return false
					}
					return true
				})
				if len(fn.Body.List) == 1 {
					var expr ast.Expr
					switch stmt := fn.Body.List[0].(type) {
					case *ast.ReturnStmt:
						if len(stmt.Results) == 1 {
							expr = stmt.Results[0]
						}
					case *ast.ExprStmt:
						expr = stmt.X
					}
					_, isCall := expr.(*ast.CallExpr)
					_, isClosure := expr.(*ast.FuncLit)
					if isCall {
						reasons = append(reasons, "only calls one level down")
					}
					if !isCall && !isClosure && len(results) == 1 && results[0] == "bool" && calls[fn.Name.Name] == 1 {
						reasons = append(reasons, "names a condition used in one place")
					}
				}
				if !fn.Name.IsExported() && branches == 0 && calls[fn.Name.Name] == 1 {
					reasons = append(reasons, "wraps statements used in one place")
				}
				if len(reasons) > 0 {
					found = append(found, Finding{Pos: fn.Pos(), Text: fmt.Sprintf("func %s: %s", fn.Name.Name, strings.Join(reasons, "; "))})
				}
			}
		}
	}
	return found
}
