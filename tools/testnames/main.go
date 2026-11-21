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

var pattern = regexp.MustCompile(`^when \S.* with \S.* then returns \S.*$`)

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
		for _, finding := range Names(file) {
			found = append(found, fmt.Sprintf("%s: %s", fset.Position(finding.Pos), finding.Text))
		}
	}
	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "every test is named `when <target> with <given conditions> then returns <expected>` (.shampoo/rules/test-names.toml):")
		fmt.Fprintln(os.Stderr, strings.Join(found, "\n"))
		os.Exit(1)
	}
}

type Finding struct {
	Pos  token.Pos
	Text string
}

func Names(file *ast.File) []Finding {
	var found []Finding
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(fn.Name.Name, "Test") || fn.Name.Name == "TestMain" {
			continue
		}
		found = append(found, inspect(fn)...)
	}
	return found
}

func inspect(fn *ast.FuncDecl) []Finding {
	var found []Finding
	literals := map[string]*ast.CompositeLit{}
	tables := map[string]ast.Expr{}
	runs := 0
	ast.Inspect(fn, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			for i, lhs := range s.Lhs {
				id, isIdent := lhs.(*ast.Ident)
				if i < len(s.Rhs) && isIdent {
					if lit, isLit := s.Rhs[i].(*ast.CompositeLit); isLit {
						literals[id.Name] = lit
					}
				}
			}
			return true
		case *ast.ValueSpec:
			for i, id := range s.Names {
				if i < len(s.Values) {
					if lit, isLit := s.Values[i].(*ast.CompositeLit); isLit {
						literals[id.Name] = lit
					}
				}
			}
			return true
		case *ast.RangeStmt:
			if v, isIdent := s.Value.(*ast.Ident); isIdent {
				tables[v.Name] = s.X
			}
			return true
		}
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
				found = append(found, Finding{Pos: name.Pos(), Text: name.Value})
			}
		case *ast.SelectorExpr:
			rows, ok := rowsOf(name, tables, literals)
			if !ok {
				found = append(found, Finding{Pos: name.Pos(), Text: "a test name is a field of a row of the table the test ranges over"})
				return true
			}
			for _, row := range rows {
				found = append(found, checkRow(row, name.Sel.Name)...)
			}
		default:
			found = append(found, Finding{Pos: name.Pos(), Text: "a test name is a string literal or a table field"})
		}
		return true
	})
	if runs == 0 {
		found = append(found, Finding{Pos: fn.Pos(), Text: fmt.Sprintf("%s names no test with t.Run", fn.Name.Name)})
	}
	return found
}

func rowsOf(name *ast.SelectorExpr, tables map[string]ast.Expr, literals map[string]*ast.CompositeLit) ([]ast.Expr, bool) {
	row, ok := name.X.(*ast.Ident)
	if !ok {
		return nil, false
	}
	switch table := tables[row.Name].(type) {
	case *ast.CompositeLit:
		return table.Elts, true
	case *ast.Ident:
		lit, ok := literals[table.Name]
		if !ok {
			return nil, false
		}
		return lit.Elts, true
	}
	return nil, false
}

func checkRow(row ast.Expr, field string) []Finding {
	if kv, ok := row.(*ast.KeyValueExpr); ok {
		row = kv.Value
	}
	lit, ok := row.(*ast.CompositeLit)
	if !ok {
		return []Finding{{Pos: row.Pos(), Text: "a table row is a composite literal"}}
	}
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != field {
			continue
		}
		value, ok := kv.Value.(*ast.BasicLit)
		if !ok {
			return []Finding{{Pos: kv.Value.Pos(), Text: "a test name is a string literal"}}
		}
		text, _ := strconv.Unquote(value.Value)
		if pattern.MatchString(text) {
			return nil
		}
		return []Finding{{Pos: value.Pos(), Text: value.Value}}
	}
	return []Finding{{Pos: lit.Pos(), Text: fmt.Sprintf("the row has no %s", field)}}
}
