//go:build small

package main

import (
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestHelpers(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		source string
		want   []string
	}{
		{name: "when Helpers with test, fuzz and example functions then returns nothing", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {}\nfunc FuzzA(f *testing.F) {}\nfunc ExampleA() {}\n", want: nil},
		{name: "when Helpers with another function then returns it", path: "x_test.go", source: "package x\nfunc build() {}\n", want: []string{"func build"}},
		{name: "when Helpers with a method named as a test then returns it", path: "x_test.go", source: "package x\nfunc (s S) TestA() {}\n", want: []string{"func TestA"}},
		{name: "when Helpers with a call of Helper then returns it", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nt.Helper()\n}\n", want: []string{"Helper()"}},
		{name: "when Helpers with a call of Helper with arguments then returns nothing", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nx.Helper(t)\n}\n", want: nil},
		{name: "when Helpers with a function literal in a test then returns nothing", path: "x_test.go", source: "package x\nfunc TestA(t *testing.T) {\nf := func() {}\nf()\n}\n", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), tt.path, tt.source, parser.SkipObjectResolution)
			if err != nil {
				t.Fatal(err)
			}

			var got []string
			for _, helper := range Helpers(file) {
				got = append(got, helper.Name)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
