//go:build small

package main

import (
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestBranches(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		source string
		want   []string
	}{
		{name: "when Branches with only statements then returns nothing", path: "x.go", source: "package x\nfunc f() {\na()\nb()\n}\n", want: nil},
		{name: "when Branches with an error check then returns nothing", path: "x.go", source: "package x\nfunc f() {\nif err != nil { return }\n}\n", want: nil},
		{name: "when Branches with an error check and its init then returns nothing", path: "x.go", source: "package x\nfunc f() {\nif err := a(); err != nil { return }\n}\n", want: nil},
		{name: "when Branches with a check of the error's kind then returns nothing", path: "x.go", source: "package x\nfunc f() {\nif err != nil && !errors.Is(err, fs.ErrNotExist) { return }\n}\n", want: nil},
		{name: "when Branches with a match of the error's type then returns nothing", path: "x.go", source: "package x\nfunc f() {\nif errors.As(err, &exit) { return }\n}\n", want: nil},
		{name: "when Branches with a type switch then returns nothing", path: "x.go", source: "package x\nfunc f() {\nswitch c := x.(type) { case A: a(c) }\n}\n", want: nil},
		{name: "when Branches with a range then returns nothing", path: "x.go", source: "package x\nfunc f() {\nfor _, x := range xs { a(x) }\n}\n", want: nil},
		{name: "when Branches with a condition on a value then returns the if", path: "x.go", source: "package x\nfunc f() {\nif len(xs) > 0 { a() }\n}\n", want: []string{"if"}},
		{name: "when Branches with an error check joined to a value then returns the if", path: "x.go", source: "package x\nfunc f() {\nif err == nil && ok { a() }\n}\n", want: []string{"if"}},
		{name: "when Branches with a switch then returns the switch", path: "x.go", source: "package x\nfunc f() {\nswitch x { case 1: a() }\n}\n", want: []string{"switch"}},
		{name: "when Branches with a conditional loop then returns the for", path: "x.go", source: "package x\nfunc f() {\nfor i := 0; i < 3; i++ { a() }\n}\n", want: []string{"for"}},
		{name: "when Branches with a select then returns the select", path: "x.go", source: "package x\nfunc f() {\nselect { case <-c: a() }\n}\n", want: []string{"select"}},
		{name: "when Branches with a condition inside an error check then returns the inner if", path: "x.go", source: "package x\nfunc f() {\nif err != nil { if ok { a() } }\n}\n", want: []string{"if"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), tt.path, tt.source, parser.SkipObjectResolution)
			if err != nil {
				t.Fatal(err)
			}

			var got []string
			for _, branch := range Branches(file) {
				got = append(got, branch.Kind)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
