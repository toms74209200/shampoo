//go:build small

package main

import (
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestNames(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		source string
		want   []string
	}{
		{name: "when Names with a subtest named by the pattern then returns nothing", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nt.Run(\"when A with b then returns c\", f)\n}\n", want: nil},
		{name: "when Names with a subtest named otherwise then returns its name", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nt.Run(\"works\", f)\n}\n", want: []string{"\"works\""}},
		{name: "when Names with a test running no subtest then returns it", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nf()\n}\n", want: []string{"TestA names no test with t.Run"}},
		{name: "when Names with TestMain then returns nothing", path: "x_test.go", source: "package x\nfunc TestMain(m *testing.M) {\nm.Run()\n}\n", want: nil},
		{name: "when Names with a table of names by the pattern then returns nothing", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\ntests := []struct{ name string }{{name: \"when A with b then returns c\"}}\nfor _, tt := range tests {\nt.Run(tt.name, f)\n}\n}\n", want: nil},
		{name: "when Names with a table row named otherwise then returns its name", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nfor _, tt := range []struct{ name string }{{name: \"works\"}} {\nt.Run(tt.name, f)\n}\n}\n", want: []string{"\"works\""}},
		{name: "when Names with a table row without the name then returns the row", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nfor _, tt := range []struct{ name string }{{}} {\nt.Run(tt.name, f)\n}\n}\n", want: []string{"the row has no name"}},
		{name: "when Names with a table row named by a variable then returns it", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nfor _, tt := range []struct{ name string }{{name: n}} {\nt.Run(tt.name, f)\n}\n}\n", want: []string{"a test name is a string literal"}},
		{name: "when Names with a table that is not a literal then returns the name", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nfor _, tt := range cases() {\nt.Run(tt.name, f)\n}\n}\n", want: []string{"a test name is a field of a row of the table the test ranges over"}},
		{name: "when Names with a subtest named by a call then returns it", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nt.Run(name(), f)\n}\n", want: []string{"a test name is a string literal or a table field"}},
		{name: "when Names with a table of rows that are not literals then returns the row", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nfor _, tt := range []row{r} {\nt.Run(tt.name, f)\n}\n}\n", want: []string{"a table row is a composite literal"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), tt.path, tt.source, parser.SkipObjectResolution)
			if err != nil {
				t.Fatal(err)
			}

			var got []string
			for _, finding := range Names(file) {
				got = append(got, finding.Text)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
