//go:build small

package hook

import (
	"reflect"
	"testing"

	"github.com/toms74209200/shampoo/internal/rule"
)

func TestStop(t *testing.T) {
	tests := []struct {
		name       string
		transition rule.Transition
		input      string
		want       Transcript
	}{
		{name: "when Stop with a task and a transcript then returns it named", transition: rule.Transition("task"), input: `{"hook_event_name":"Stop","transcript_path":"/tmp/t.jsonl"}`, want: Named{Path: "/tmp/t.jsonl"}},
		{name: "when Stop with an edit and a transcript then returns none", transition: rule.Transition("edit"), input: `{"transcript_path":"/tmp/t.jsonl"}`, want: Unnamed{}},
		{name: "when Stop with a task and no transcript then returns none", transition: rule.Transition("task"), input: `{"hook_event_name":"Stop"}`, want: Unnamed{}},
		{name: "when Stop with a task and no input then returns none", transition: rule.Transition("task"), input: "", want: Unnamed{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Stop(tt.transition, []byte(tt.input))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
