//go:build small

package rule

import (
	"crypto/rand"
	"reflect"
	"regexp"
	"testing"
)

func TestSelect(t *testing.T) {
	goFiles := Checked{ID: rand.Text(), Transitions: []Transition{Transition("edit"), Transition("pr")}, Paths: Paths{patterns: []*regexp.Regexp{regexp.MustCompile(`^.*\.go$`)}}}
	anyFile := Checked{ID: rand.Text(), Transitions: []Transition{Transition("edit")}}
	prOnly := Checked{ID: rand.Text(), Transitions: []Transition{Transition("pr")}}
	unchecked := Unchecked{ID: rand.Text()}
	tests := []struct {
		name       string
		transition Transition
		changed    []string
		want       []Checked
	}{
		{name: "when Select with a transition and changed paths covered by its rules then returns those rules", transition: Transition("edit"), changed: []string{"main.go"}, want: []Checked{goFiles, anyFile}},
		{name: "when Select with changed paths outside a rule's paths then returns the rules without it", transition: Transition("edit"), changed: []string{"README.md"}, want: []Checked{anyFile}},
		{name: "when Select with a transition no rule guards then returns no rule", transition: Transition("milestone"), changed: []string{"main.go"}, want: nil},
		{name: "when Select with no changed path then returns no rule", transition: Transition("edit"), changed: nil, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Select([]Rule{goFiles, unchecked, anyFile, prOnly}, tt.transition, tt.changed)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
