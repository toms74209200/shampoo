//go:build small

package main

import (
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestProhibited(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		source string
		want   []string
	}{
		{name: "when Prohibited with a deep parse function then returns nothing", path: "a/a.go", source: "package a\nfunc Parse(data []byte) (Port, error) {\nif len(data) == 0 { return 0, nil }\nreturn 1, nil\n}\n", want: nil},
		{name: "when Prohibited with main and init then returns nothing", path: "a/a.go", source: "package a\nfunc main() { run() }\nfunc init() { run() }\n", want: nil},
		{name: "when Prohibited with primitives into primitives then returns the reason", path: "a/a.go", source: "package a\nfunc Double(n int) int {\nif n > 0 { return n * 2 }\nreturn 0\n}\n", want: []string{"func Double: converts primitives into primitives"}},
		{name: "when Prohibited with the type it was given returned then returns the reason", path: "a/a.go", source: "package a\nfunc Trim(p Port) Port {\nif p > 0 { return p }\nreturn 0\n}\n", want: []string{"func Trim: returns the type it was given: Port"}},
		{name: "when Prohibited with one call one level down then returns the reason", path: "a/a.go", source: "package a\nfunc Load(p Path) (Data, error) {\nreturn read(p)\n}\n", want: []string{"func Load: only calls one level down"}},
		{name: "when Prohibited with a returned closure then returns nothing", path: "a/a.go", source: "package a\nfunc Retrying(f F) G {\nreturn func() { f() }\n}\n", want: nil},
		{name: "when Prohibited with a condition used in one place then returns the reason", path: "a/a.go", source: "package a\nfunc Ready(s State) bool {\nreturn s.ok && s.done\n}\nfunc Run(s State) (Out, error) {\nif Ready(s) { return Out{}, nil }\nreturn Out{}, nil\n}\n", want: []string{"func Ready: names a condition used in one place"}},
		{name: "when Prohibited with unexported statements used in one place then returns the reason", path: "a/a.go", source: "package a\nfunc save(s State) {\na(s)\nb(s)\n}\nfunc Run(s State) (Out, error) {\nif s.ok { save(s) }\nreturn Out{}, nil\n}\n", want: []string{"func save: wraps statements used in one place"}},
		{name: "when Prohibited with unexported statements used in two places then returns nothing", path: "a/a.go", source: "package a\nfunc save(s State) {\na(s)\nb(s)\n}\nfunc Run(s State) (Out, error) {\nif s.ok { save(s) }\nsave(s)\nreturn Out{}, nil\n}\n", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), tt.path, tt.source, parser.SkipObjectResolution)
			if err != nil {
				t.Fatal(err)
			}

			var got []string
			for _, finding := range Prohibited([]Source{{Path: tt.path, File: file}}) {
				got = append(got, finding.Text)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
