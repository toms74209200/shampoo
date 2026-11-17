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
	directives := []string{"//go:", "//nolint:", "//line "}
	fset := token.NewFileSet()
	var found []string
	paths := bufio.NewScanner(os.Stdin)
	for paths.Scan() {
		file, err := parser.ParseFile(fset, paths.Text(), nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if ast.IsGenerated(file) {
			continue
		}
		for _, group := range file.Comments {
			for _, c := range group.List {
				if !slices.ContainsFunc(directives, func(d string) bool { return strings.HasPrefix(c.Text, d) }) {
					found = append(found, fmt.Sprintf("%s: %s", fset.Position(c.Pos()), c.Text))
				}
			}
		}
	}
	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "comments are prohibited (.shampoo/rules/comments.toml):")
		fmt.Fprintln(os.Stderr, strings.Join(found, "\n"))
		os.Exit(1)
	}
}
