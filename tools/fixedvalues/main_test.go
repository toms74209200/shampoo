//go:build small

package main

import (
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestLiterals(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		source string
		want   []string
	}{
		{name: "when Literals with a literal in a slice then returns nothing", path: "x_test.go", source: "package x\nvar a = []int{1, 2}\n", want: nil},
		{name: "when Literals with a literal in a map then returns nothing", path: "x_test.go", source: "package x\nvar a = map[string]int{\"k\": 1}\n", want: nil},
		{name: "when Literals with a literal in a table of struct cases then returns nothing", path: "x_test.go", source: "package x\nvar a = []struct{ n int }{{n: 1}}\n", want: nil},
		{name: "when Literals with a literal in a struct outside a table then returns it", path: "x_test.go", source: "package x\nvar a = S{n: 1}\n", want: []string{"1"}},
		{name: "when Literals with a literal in an assertion then returns it", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nif got != 3 { t.Fail() }\n}\n", want: []string{"3"}},
		{name: "when Literals with a literal in a failure message then returns nothing", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nt.Errorf(\"got %v\", got)\n}\n", want: nil},
		{name: "when Literals with the name of a subtest then returns nothing", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nt.Run(\"a\", f)\n}\n", want: nil},
		{name: "when Literals with a literal as the body of a subtest then returns it", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nt.Run(name, g(\"a\"))\n}\n", want: []string{"\"a\""}},
		{name: "when Literals with an import path then returns nothing", path: "x_test.go", source: "package x\nimport \"testing\"\n", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), tt.path, tt.source, parser.SkipObjectResolution)
			if err != nil {
				t.Fatal(err)
			}

			var got []string
			for _, lit := range Literals(file) {
				got = append(got, lit.Value)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
