//go:build small

package rule

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	id := strings.ToLower(rand.Text())
	check := strings.ToLower(rand.Text())
	statement := rand.Text()
	command := rand.Text()
	reason := rand.Text()
	tests := []struct {
		name    string
		id      string
		data    string
		want    Rule
		wantErr bool
	}{
		{
			name: "when Parse with a rule judging its check's exit code then returns the checked rule",
			id:   id,
			data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\", \"pr\"]\npaths = [\"**.go\", \"tools/*.go\", \"a?b\"]\njudge = { exit = { 1 = \"task-failure\", 2 = \"workflow-failure\", 3 = \"budget\", other = \"workflow-failure\" } }\nbudget = { violations = 2, runs = 20 }\n", statement, check),
			want: Checked{
				ID:          id,
				Statement:   statement,
				Check:       check,
				Judge:       ByExit{Exit: Exit{codes: map[int]Outcome{1: Outcome("task-failure"), 2: Outcome("workflow-failure"), 3: Outcome("budget")}, other: Outcome("workflow-failure")}},
				Transitions: []Transition{Transition("edit"), Transition("pr")},
				Paths:       Paths{patterns: []*regexp.Regexp{regexp.MustCompile(`^.*\.go$`), regexp.MustCompile(`^tools/[^/]*\.go$`), regexp.MustCompile(`^a[^/]?b$`)}},
				Budget:      map[Transition]Count{Transition("edit"): {Violations: 2, Runs: 20}, Transition("pr"): {Violations: 2, Runs: 20}},
				Stop:        map[Transition]Count{Transition("edit"): {Violations: 8, Runs: 10}, Transition("pr"): {Violations: 2, Runs: 5}},
			},
		},
		{
			name: "when Parse with a judge command a zero budget and a stop then returns the checked rule judged by the command",
			id:   id,
			data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"task\"]\njudge = { command = %q, exit = { other = \"budget\" } }\nbudget = { violations = 0, runs = 1 }\nstop = { violations = 1, runs = 3 }\n", statement, check, command),
			want: Checked{
				ID:          id,
				Statement:   statement,
				Check:       check,
				Judge:       ByCommand{Command: command, Exit: Exit{codes: map[int]Outcome{}, other: Outcome("budget")}},
				Transitions: []Transition{Transition("task")},
				Budget:      map[Transition]Count{Transition("task"): {Violations: 0, Runs: 1}},
				Stop:        map[Transition]Count{Transition("task"): {Violations: 1, Runs: 3}},
			},
		},
		{
			name: "when Parse with a rule without a budget then returns the checked rule with the default budget",
			id:   id,
			data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\n", statement, check),
			want: Checked{
				ID:          id,
				Statement:   statement,
				Check:       check,
				Judge:       ByExit{Exit: Exit{codes: map[int]Outcome{}, other: Outcome("task-failure")}},
				Transitions: []Transition{Transition("pr")},
				Budget:      map[Transition]Count{Transition("pr"): {Violations: 2, Runs: 20}},
				Stop:        map[Transition]Count{Transition("pr"): {Violations: 2, Runs: 5}},
			},
		},
		{
			name: "when Parse with budgets by transition then returns them with the defaults for the transitions left out",
			id:   id,
			data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\", \"task\", \"pr\"]\njudge = { exit = { other = \"task-failure\" } }\n[budget]\ntask = { violations = 1, runs = 4 }\n[stop]\nedit = { violations = 3, runs = 3 }\n", statement, check),
			want: Checked{
				ID:          id,
				Statement:   statement,
				Check:       check,
				Judge:       ByExit{Exit: Exit{codes: map[int]Outcome{}, other: Outcome("task-failure")}},
				Transitions: []Transition{Transition("edit"), Transition("task"), Transition("pr")},
				Budget:      map[Transition]Count{Transition("edit"): {Violations: 10, Runs: 20}, Transition("task"): {Violations: 1, Runs: 4}, Transition("pr"): {Violations: 2, Runs: 20}},
				Stop:        map[Transition]Count{Transition("edit"): {Violations: 3, Runs: 3}, Transition("task"): {Violations: 2, Runs: 5}, Transition("pr"): {Violations: 2, Runs: 5}},
			},
		},
		{
			name: "when Parse with a rule and its reason for no check then returns the unchecked rule",
			id:   id,
			data: fmt.Sprintf("rule = %q\nunchecked = %q\n", statement, reason),
			want: Unchecked{ID: id, Statement: statement, Reason: reason},
		},
		{name: "when Parse with a rule without a judge then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\n", statement, check), wantErr: true},
		{name: "when Parse with an empty judge command then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { command = \"\", exit = { other = \"task-failure\" } }\n", statement, check), wantErr: true},
		{name: "when Parse with a judge whose exit has no other then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { 1 = \"task-failure\" } }\n", statement, check), wantErr: true},
		{name: "when Parse with a judge mapping an exit code to pass then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"pass\" } }\n", statement, check), wantErr: true},
		{name: "when Parse with a judge with a non-numeric exit code then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { x = \"task-failure\", other = \"task-failure\" } }\n", statement, check), wantErr: true},
		{name: "when Parse with a judge with exit code zero then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { 0 = \"task-failure\", other = \"task-failure\" } }\n", statement, check), wantErr: true},
		{name: "when Parse with a budget of zero runs then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\nbudget = { violations = 0, runs = 0 }\n", statement, check), wantErr: true},
		{name: "when Parse with a budget of negative violations then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\nbudget = { violations = -1, runs = 2 }\n", statement, check), wantErr: true},
		{name: "when Parse with a budget of more violations than runs then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\nbudget = { violations = 3, runs = 2 }\n", statement, check), wantErr: true},
		{name: "when Parse with a stop at zero violations then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\nstop = { violations = 0, runs = 2 }\n", statement, check), wantErr: true},
		{name: "when Parse with a stop of more violations than runs then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\nstop = { violations = 3, runs = 2 }\n", statement, check), wantErr: true},
		{name: "when Parse with a budget written as a share then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\nbudget = 0.1\n", statement, check), wantErr: true},
		{name: "when Parse with a budget for a transition the rule does not guard then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\n[budget]\nedit = { violations = 1, runs = 4 }\n", statement, check), wantErr: true},
		{name: "when Parse with a budget by transition that is not a count then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\n[budget]\npr = 3\n", statement, check), wantErr: true},
		{name: "when Parse with a stop by transition of zero violations then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\n[stop]\npr = { violations = 0, runs = 2 }\n", statement, check), wantErr: true},
		{name: "when Parse with a count missing its runs then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\nbudget = { violations = 1 }\n", statement, check), wantErr: true},
		{name: "when Parse with a reason and a stop then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = %q\nstop = { violations = 1, runs = 1 }\n", statement, reason), wantErr: true},
		{name: "when Parse with a reason and a budget then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = %q\nbudget = { violations = 1, runs = 2 }\n", statement, reason), wantErr: true},
		{name: "when Parse with an id outside lowercase words then returns an error", id: "Bad_ID", data: fmt.Sprintf("rule = %q\nunchecked = %q\n", statement, reason), wantErr: true},
		{name: "when Parse with invalid TOML then returns an error", id: id, data: "rule = ", wantErr: true},
		{name: "when Parse with an unknown field then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = %q\nnote = %q\n", statement, reason, reason), wantErr: true},
		{name: "when Parse with no rule then returns an error", id: id, data: fmt.Sprintf("unchecked = %q\n", reason), wantErr: true},
		{name: "when Parse with a reason and a check then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = %q\ncheck = %q\n", statement, reason, check), wantErr: true},
		{name: "when Parse with an empty reason then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = \"\"\n", statement), wantErr: true},
		{name: "when Parse with neither a check nor a reason then returns an error", id: id, data: fmt.Sprintf("rule = %q\n", statement), wantErr: true},
		{name: "when Parse with a check id outside lowercase words then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = \"Bad_ID\"\ntransitions = [\"edit\"]\njudge = { exit = { other = \"task-failure\" } }\n", statement), wantErr: true},
		{name: "when Parse with a check without transitions then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\njudge = { exit = { other = \"task-failure\" } }\n", statement, check), wantErr: true},
		{name: "when Parse with an unknown transition then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"merge\"]\njudge = { exit = { other = \"task-failure\" } }\n", statement, check), wantErr: true},
		{name: "when Parse with an unsupported path pattern then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\npaths = [\"[ab].go\"]\njudge = { exit = { other = \"task-failure\" } }\n", statement, check), wantErr: true},
		{name: "when Parse with an empty path pattern then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\npaths = [\"\"]\njudge = { exit = { other = \"task-failure\" } }\n", statement, check), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.id, []byte(tt.data))

			if tt.wantErr && (err == nil || got != nil) {
				t.Errorf("got %v, %v, want an error", got, err)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
