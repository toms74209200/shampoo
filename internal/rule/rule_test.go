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
			name: "when Parse with a rule and its check then returns the checked rule",
			id:   id,
			data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\", \"pr\"]\npaths = [\"**.go\", \"tools/*.go\", \"a?b\"]\nsize = \"medium\"\nexit = { 1 = \"task-failure\", 2 = \"workflow-failure\", 3 = \"structure-budget\", other = \"workflow-failure\" }\n", statement, command),
			want: Checked{
				ID:          id,
				Statement:   statement,
				Command:     command,
				Transitions: []Transition{Transition("edit"), Transition("pr")},
				Paths:       Paths{patterns: []*regexp.Regexp{regexp.MustCompile(`^.*\.go$`), regexp.MustCompile(`^tools/[^/]*\.go$`), regexp.MustCompile(`^a[^/]?b$`)}},
				Size:        Size("medium"),
				Exit:        Exit{codes: map[int]Outcome{1: Outcome("task-failure"), 2: Outcome("workflow-failure"), 3: Outcome("structure-budget")}, other: Outcome("workflow-failure")},
			},
		},
		{
			name: "when Parse with a rule without paths then returns the checked rule covering every path",
			id:   id,
			data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"task\", \"milestone\"]\nsize = \"large\"\nexit = { other = \"task-failure\" }\n", statement, command),
			want: Checked{
				ID:          id,
				Statement:   statement,
				Command:     command,
				Transitions: []Transition{Transition("task"), Transition("milestone")},
				Size:        Size("large"),
				Exit:        Exit{codes: map[int]Outcome{}, other: Outcome("task-failure")},
			},
		},
		{
			name: "when Parse with a rule and its reason for no check then returns the unchecked rule",
			id:   id,
			data: fmt.Sprintf("rule = %q\nunchecked = %q\n", statement, reason),
			want: Unchecked{ID: id, Statement: statement, Reason: reason},
		},
		{name: "when Parse with an id outside lowercase words then returns an error", id: "Bad_ID", data: fmt.Sprintf("rule = %q\nunchecked = %q\n", statement, reason), wantErr: true},
		{name: "when Parse with invalid TOML then returns an error", id: id, data: "rule = ", wantErr: true},
		{name: "when Parse with an unknown field then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = %q\nnote = %q\n", statement, reason, reason), wantErr: true},
		{name: "when Parse with no rule then returns an error", id: id, data: fmt.Sprintf("unchecked = %q\n", reason), wantErr: true},
		{name: "when Parse with a reason and a check then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = %q\ncheck = %q\n", statement, reason, command), wantErr: true},
		{name: "when Parse with an empty reason then returns an error", id: id, data: fmt.Sprintf("rule = %q\nunchecked = \"\"\n", statement), wantErr: true},
		{name: "when Parse with neither a check nor a reason then returns an error", id: id, data: fmt.Sprintf("rule = %q\n", statement), wantErr: true},
		{name: "when Parse with a check without transitions then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\nsize = \"small\"\nexit = { other = \"task-failure\" }\n", statement, command), wantErr: true},
		{name: "when Parse with an unknown transition then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"merge\"]\nsize = \"small\"\nexit = { other = \"task-failure\" }\n", statement, command), wantErr: true},
		{name: "when Parse with an unknown size then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\nsize = \"huge\"\nexit = { other = \"task-failure\" }\n", statement, command), wantErr: true},
		{name: "when Parse with an unsupported path pattern then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\npaths = [\"[ab].go\"]\nsize = \"small\"\nexit = { other = \"task-failure\" }\n", statement, command), wantErr: true},
		{name: "when Parse with an empty path pattern then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\npaths = [\"\"]\nsize = \"small\"\nexit = { other = \"task-failure\" }\n", statement, command), wantErr: true},
		{name: "when Parse with an unknown exit outcome then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\nsize = \"small\"\nexit = { other = \"pass\" }\n", statement, command), wantErr: true},
		{name: "when Parse with a non-numeric exit code then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\nsize = \"small\"\nexit = { x = \"task-failure\", other = \"task-failure\" }\n", statement, command), wantErr: true},
		{name: "when Parse with exit code zero then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\nsize = \"small\"\nexit = { 0 = \"task-failure\", other = \"task-failure\" }\n", statement, command), wantErr: true},
		{name: "when Parse with exit without other then returns an error", id: id, data: fmt.Sprintf("rule = %q\ncheck = %q\ntransitions = [\"edit\"]\nsize = \"small\"\nexit = { 1 = \"task-failure\" }\n", statement, command), wantErr: true},
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
