//go:build small

package main

import (
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestUses(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{name: "when Uses with a test using its target only then returns nothing", files: map[string]string{"a/a.go": "package a\nfunc Parse() {}\nfunc Other() {}\n", "a/a_test.go": "package a\nfunc TestParse(t *testing.T) {\nt.Run(\"when Parse with x then returns y\", func(t *testing.T) { Parse() })\n}\n"}, want: nil},
		{name: "when Uses with a test using another function of the package then returns it", files: map[string]string{"a/a.go": "package a\nfunc Parse() {}\nfunc Other() {}\n", "a/a_test.go": "package a\nfunc TestParse(t *testing.T) {\nt.Run(\"when Parse with x then returns y\", func(t *testing.T) { Other() })\n}\n"}, want: []string{"TestParse uses Other"}},
		{name: "when Uses with a test using a variable of the package then returns it", files: map[string]string{"a/a.go": "package a\nfunc Parse() {}\nvar limit = 3\n", "a/a_test.go": "package a\nfunc TestParse(t *testing.T) {\nt.Run(\"when Parse with x then returns y\", func(t *testing.T) { Parse(); _ = limit })\n}\n"}, want: []string{"TestParse uses limit"}},
		{name: "when Uses with a test using a type of the package then returns nothing", files: map[string]string{"a/a.go": "package a\ntype Input struct{}\nfunc Parse(Input) {}\n", "a/a_test.go": "package a\nfunc TestParse(t *testing.T) {\nt.Run(\"when Parse with x then returns y\", func(t *testing.T) { Parse(Input{}) })\n}\n"}, want: nil},
		{name: "when Uses with a field named as a function of the package then returns nothing", files: map[string]string{"a/a.go": "package a\nfunc Parse() {}\nfunc Other() {}\n", "a/a_test.go": "package a\nfunc TestParse(t *testing.T) {\nt.Run(\"when Parse with x then returns y\", func(t *testing.T) { Parse(); _ = s.Other; _ = S{Other: 1} })\n}\n"}, want: nil},
		{name: "when Uses with a function of another package then returns nothing", files: map[string]string{"a/a.go": "package a\nfunc Parse() {}\n", "b/b.go": "package b\nfunc Other() {}\n", "a/a_test.go": "package a\nfunc TestParse(t *testing.T) {\nt.Run(\"when Parse with x then returns y\", func(t *testing.T) { Parse(); Other() })\n}\n"}, want: nil},
		{name: "when Uses with a black-box test then returns nothing", files: map[string]string{"a/a.go": "package a\nfunc Parse() {}\nfunc Other() {}\n", "a/a_test.go": "package a_test\nfunc TestParse(t *testing.T) {\nt.Run(\"when Parse with x then returns y\", func(t *testing.T) { Other() })\n}\n"}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sources []Source
			for path, source := range tt.files {
				file, err := parser.ParseFile(token.NewFileSet(), path, source, parser.SkipObjectResolution)
				if err != nil {
					t.Fatal(err)
				}
				sources = append(sources, Source{Path: path, File: file})
			}

			var got []string
			for _, use := range Uses(sources) {
				got = append(got, use.Text)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
