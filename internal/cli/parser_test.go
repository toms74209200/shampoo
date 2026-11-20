//go:build small

package cli

import (
	"reflect"
	"testing"

	"github.com/toms74209200/shampoo/internal/rule"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Command
		wantErr bool
	}{
		{name: "when Parse with guard and edit then returns guarding edit", args: []string{"guard", "edit"}, want: Guard{Transition: rule.Transition("edit")}},
		{name: "when Parse with guard and task then returns guarding task", args: []string{"guard", "task"}, want: Guard{Transition: rule.Transition("task")}},
		{name: "when Parse with next then returns next", args: []string{"next"}, want: Next{}},
		{name: "when Parse with review then returns review", args: []string{"review"}, want: Review{}},
		{name: "when Parse with policies then returns policies", args: []string{"policies"}, want: Policies{}},
		{name: "when Parse with ask then returns ask", args: []string{"ask"}, want: Ask{}},
		{name: "when Parse with guard and a transition checked by CI then returns an error", args: []string{"guard", "pr"}, wantErr: true},
		{name: "when Parse with guard and an unknown transition then returns an error", args: []string{"guard", "merge"}, wantErr: true},
		{name: "when Parse with guard and no transition then returns an error", args: []string{"guard"}, wantErr: true},
		{name: "when Parse with next and an extra argument then returns an error", args: []string{"next", "now"}, wantErr: true},
		{name: "when Parse with an unknown command then returns an error", args: []string{"run"}, wantErr: true},
		{name: "when Parse with no arguments then returns an error", args: []string{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)

			if tt.wantErr && err == nil {
				t.Errorf("got %#v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
