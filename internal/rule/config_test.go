//go:build small

package rule

import (
	"crypto/rand"
	"reflect"
	"regexp"
	"testing"
)

func TestLink(t *testing.T) {
	build := Check{ID: rand.Text()}
	test := Check{ID: rand.Text(), Needs: []string{build.ID}}
	rules := []Rule{Checked{ID: rand.Text(), Check: test.ID}, Unchecked{ID: rand.Text()}}
	tests := []struct {
		name    string
		rules   []Rule
		checks  []Check
		want    Config
		wantErr bool
	}{
		{
			name:   "when Link with rules pointing to checks that need defined checks then returns the config",
			rules:  rules,
			checks: []Check{test, build},
			want:   Config{rules: rules, checks: map[string]Check{test.ID: test, build.ID: build}},
		},
		{name: "when Link with a rule pointing to an undefined check then returns an error", rules: []Rule{Checked{ID: rand.Text(), Check: rand.Text()}}, checks: []Check{build}, wantErr: true},
		{name: "when Link with a check needing an undefined check then returns an error", checks: []Check{test}, wantErr: true},
		{name: "when Link with a check defined twice then returns an error", checks: []Check{build, build}, wantErr: true},
		{name: "when Link with checks needing each other then returns an error", checks: []Check{{ID: build.ID, Needs: []string{test.ID}}, test}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Link(tt.rules, tt.checks)

			if tt.wantErr && err == nil {
				t.Errorf("got %v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}

func TestPlan(t *testing.T) {
	build := Check{ID: rand.Text()}
	acceptance := Check{ID: rand.Text(), Needs: []string{build.ID}}
	lint := Check{ID: rand.Text(), Needs: []string{build.ID}}
	goFiles := Checked{ID: rand.Text(), Check: lint.ID, Transitions: []Transition{Transition("edit"), Transition("pr")}, Paths: Paths{patterns: []*regexp.Regexp{regexp.MustCompile(`^.*\.go$`)}}}
	mocks := Checked{ID: rand.Text(), Check: lint.ID, Transitions: []Transition{Transition("edit")}}
	regressions := Checked{ID: rand.Text(), Check: acceptance.ID, Transitions: []Transition{Transition("edit")}}
	prOnly := Checked{ID: rand.Text(), Check: build.ID, Transitions: []Transition{Transition("pr")}}
	config := Config{
		rules:  []Rule{goFiles, Unchecked{ID: rand.Text()}, mocks, regressions, prOnly},
		checks: map[string]Check{build.ID: build, acceptance.ID: acceptance, lint.ID: lint},
	}
	tests := []struct {
		name       string
		transition Transition
		changed    []string
		wantRules  []Checked
		wantChecks []Check
	}{
		{name: "when Plan with rules sharing a check then returns the check once after the checks it needs", transition: Transition("edit"), changed: []string{"main.go"}, wantRules: []Checked{goFiles, mocks, regressions}, wantChecks: []Check{build, lint, acceptance}},
		{name: "when Plan with changed paths outside a rule's paths then returns the rules without it", transition: Transition("pr"), changed: []string{"README.md"}, wantRules: []Checked{prOnly}, wantChecks: []Check{build}},
		{name: "when Plan with a transition no rule guards then returns no rule and no check", transition: Transition("milestone"), changed: []string{"main.go"}},
		{name: "when Plan with no changed path then returns no rule and no check", transition: Transition("edit"), changed: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRules, gotChecks := config.Plan(tt.transition, tt.changed)

			if !reflect.DeepEqual(gotRules, tt.wantRules) || !reflect.DeepEqual(gotChecks, tt.wantChecks) {
				t.Errorf("got %#v, %#v, want %#v, %#v", gotRules, gotChecks, tt.wantRules, tt.wantChecks)
			}
		})
	}
}

func TestWindow(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   int
	}{
		{name: "when Window with rules of several counts then returns the most runs of any budget or stop", config: Config{rules: []Rule{Checked{Transitions: []Transition{"edit", "task"}, Budget: map[Transition]Count{"edit": {Violations: 10, Runs: 20}, "task": {Violations: 2, Runs: 30}}, Stop: map[Transition]Count{"edit": {Violations: 8, Runs: 10}, "task": {Violations: 2, Runs: 5}}}, Unchecked{}, Checked{Transitions: []Transition{"pr"}, Budget: map[Transition]Count{"pr": {Violations: 1, Runs: 4}}, Stop: map[Transition]Count{"pr": {Violations: 1, Runs: 40}}}}}, want: 40},
		{name: "when Window with no checked rule then returns zero", config: Config{rules: []Rule{Unchecked{}}}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.Window()

			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
