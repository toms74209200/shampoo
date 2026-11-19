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
			name: "when Parse with a rule pointing to its check then returns the checked rule judged by the check",
			id:   id,
			data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\", \"pr\"]\npaths = [\"**.go\", \"tools/*.go\", \"a?b\"]\n", statement, check),
			want: Checked{
				ID:          id,
				Statement:   statement,
				Check:       check,
				Judge:       ByCheck{},
				Transitions: []Transition{Transition("edit"), Transition("pr")},
				Paths:       Paths{patterns: []*regexp.Regexp{regexp.MustCompile(`^.*\.go$`), regexp.MustCompile(`^tools/[^/]*\.go$`), regexp.MustCompile(`^a[^/]?b$`)}},
			},
		},
		{
			name: "when Parse with a judge charging the budget then returns the checked rule with its judge and budget",
			id:   id,
			data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { command = %q, exit = { 1 = \"task-failure\", 2 = \"workflow-failure\", 3 = \"budget\", other = \"workflow-failure\" } }\nbudget = 0.1\n", statement, check, command),
			want: Checked{
				ID:          id,
				Statement:   statement,
				Check:       check,
				Judge:       Judged{Command: command, Exit: Exit{codes: map[int]Outcome{1: Outcome("task-failure"), 2: Outcome("workflow-failure"), 3: Outcome("budget")}, other: Outcome("workflow-failure")}},
				Transitions: []Transition{Transition("pr")},
				Budget:      0.1,
			},
		},
		{
			name: "when Parse with an integer budget then returns the checked rule with that budget",
			id:   id,
			data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"task\"]\njudge = { command = %q, exit = { other = \"budget\" } }\nbudget = 1\n", statement, check, command),
			want: Checked{
				ID:          id,
				Statement:   statement,
				Check:       check,
				Judge:       Judged{Command: command, Exit: Exit{codes: map[int]Outcome{}, other: Outcome("budget")}},
				Transitions: []Transition{Transition("task")},
				Budget:      1,
			},
		},
		{
			name: "when Parse with a rule and its reason for no check then returns the unchecked rule",
			id:   id,
			data: fmt.Sprintf("rule = %q\nunchecked = %q\n", statement, reason),
			want: Unchecked{ID: id, Statement: statement, Reason: reason},
		},
		{name: "when Parse with a judge charging the budget without a budget then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { command = %q, exit = { 3 = \"budget\", other = \"task-failure\" } }\n", statement, check, command), wantErr: true},
		{name: "when Parse with a judge without a command then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { exit = { other = \"task-failure\" } }\n", statement, check), wantErr: true},
		{name: "when Parse with a judge whose exit has no other then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\njudge = { command = %q, exit = { 1 = \"task-failure\" } }\n", statement, check, command), wantErr: true},
		{name: "when Parse with a budget above one then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\nbudget = 1.5\n", statement, check), wantErr: true},
		{name: "when Parse with a budget that is not a number then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"pr\"]\nbudget = \"low\"\n", statement, check), wantErr: true},
		{name: "when Parse with a reason and a budget then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = %q\nbudget = 0.1\n", statement, reason), wantErr: true},
		{name: "when Parse with an id outside lowercase words then returns an error", id: "Bad_ID", data: fmt.Sprintf("rule = %q\nunchecked = %q\n", statement, reason), wantErr: true},
		{name: "when Parse with invalid TOML then returns an error", id: id, data: "rule = ", wantErr: true},
		{name: "when Parse with an unknown field then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = %q\nnote = %q\n", statement, reason, reason), wantErr: true},
		{name: "when Parse with no rule then returns an error", id: id, data: fmt.Sprintf("unchecked = %q\n", reason), wantErr: true},
		{name: "when Parse with a reason and a check then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = %q\ncheck = %q\n", statement, reason, check), wantErr: true},
		{name: "when Parse with an empty reason then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = \"\"\n", statement), wantErr: true},
		{name: "when Parse with neither a check nor a reason then returns an error", id: id, data: fmt.Sprintf("rule = %q\n", statement), wantErr: true},
		{name: "when Parse with a check id outside lowercase words then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = \"Bad_ID\"\ntransitions = [\"edit\"]\n", statement), wantErr: true},
		{name: "when Parse with a check without transitions then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\n", statement, check), wantErr: true},
		{name: "when Parse with an unknown transition then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"merge\"]\n", statement, check), wantErr: true},
		{name: "when Parse with an unsupported path pattern then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\npaths = [\"[ab].go\"]\n", statement, check), wantErr: true},
		{name: "when Parse with an empty path pattern then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\npaths = [\"\"]\n", statement, check), wantErr: true},
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
