//go:build small

package policy

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		data    string
		want    Policy
		wantErr bool
	}{
		{name: "when Parse with a text and an example then returns the policy not for review", id: "layout", data: "text = \"a\"\nexample = \"b\"\n", want: Policy{ID: "layout", Text: "a", Example: "b"}},
		{name: "when Parse with review then returns the policy for review", id: "functions", data: "text = \"a\"\nreview = true\n", want: Policy{ID: "functions", Text: "a", Review: true}},
		{name: "when Parse with an invalid id then returns an error", id: "Bad_ID", data: "text = \"a\"\n", wantErr: true},
		{name: "when Parse with invalid TOML then returns an error", id: "layout", data: "text = ", wantErr: true},
		{name: "when Parse with an unknown field then returns an error", id: "layout", data: "text = \"a\"\nreviewed = true\n", wantErr: true},
		{name: "when Parse with no text then returns an error", id: "layout", data: "review = true\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.id, []byte(tt.data))

			if tt.wantErr && err == nil {
				t.Errorf("got %#v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		policies []Policy
		want     string
	}{
		{name: "when Render with policies then returns each with its text and its example", policies: []Policy{{ID: "layout", Text: "\nmain.go runs.\n", Example: "\nmain.go\n"}, {ID: "functions", Text: "No wrappers.", Review: true}}, want: "# Policies of this project (.shampoo/policies/)\n\n## layout\n\nmain.go runs.\n\nmain.go\n\n## functions\n\nNo wrappers.\n"},
		{name: "when Render with no policies then returns nothing", policies: nil, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.policies)

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
