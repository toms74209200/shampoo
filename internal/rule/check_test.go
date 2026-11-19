//go:build small

package rule

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseCheck(t *testing.T) {
	id := strings.ToLower(rand.Text())
	need := strings.ToLower(rand.Text())
	command := rand.Text()
	tests := []struct {
		name    string
		id      string
		data    string
		want    Check
		wantErr bool
	}{
		{
			name: "when ParseCheck with a command its size its exit and its needs then returns the check",
			id:   id,
			data: fmt.Sprintf("command = %q\nsize = \"medium\"\nexit = { 1 = \"task-failure\", other = \"workflow-failure\" }\nneeds = [%q]\n", command, need),
			want: Check{ID: id, Command: command, Size: Size("medium"), Exit: Exit{codes: map[int]Outcome{1: Outcome("task-failure")}, other: Outcome("workflow-failure")}, Needs: []string{need}},
		},
		{
			name: "when ParseCheck with no needs then returns the check needing nothing",
			id:   id,
			data: fmt.Sprintf("command = %q\nsize = \"large\"\nexit = { other = \"task-failure\" }\n", command),
			want: Check{ID: id, Command: command, Size: Size("large"), Exit: Exit{codes: map[int]Outcome{}, other: Outcome("task-failure")}},
		},
		{name: "when ParseCheck with an id outside lowercase words then returns an error", id: "Bad_ID", data: fmt.Sprintf("command = %q\nsize = \"small\"\nexit = { other = \"task-failure\" }\n", command), wantErr: true},
		{name: "when ParseCheck with invalid TOML then returns an error", id: id, data: "command = ", wantErr: true},
		{name: "when ParseCheck with an unknown field then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nexit = { other = \"task-failure\" }\ntransitions = [\"edit\"]\n", command), wantErr: true},
		{name: "when ParseCheck with no command then returns an error", id: id, data: "size = \"small\"\nexit = { other = \"task-failure\" }\n", wantErr: true},
		{name: "when ParseCheck with an unknown size then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"huge\"\nexit = { other = \"task-failure\" }\n", command), wantErr: true},
		{name: "when ParseCheck with an exit charging a budget then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nexit = { 3 = \"budget\", other = \"task-failure\" }\n", command), wantErr: true},
		{name: "when ParseCheck with an exit mapped to pass then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nexit = { other = \"pass\" }\n", command), wantErr: true},
		{name: "when ParseCheck with a non-numeric exit code then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nexit = { x = \"task-failure\", other = \"task-failure\" }\n", command), wantErr: true},
		{name: "when ParseCheck with exit code zero then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nexit = { 0 = \"task-failure\", other = \"task-failure\" }\n", command), wantErr: true},
		{name: "when ParseCheck with exit without other then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nexit = { 1 = \"task-failure\" }\n", command), wantErr: true},
		{name: "when ParseCheck with a need of itself then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nexit = { other = \"task-failure\" }\nneeds = [%q]\n", command, id), wantErr: true},
		{name: "when ParseCheck with a need outside lowercase words then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nexit = { other = \"task-failure\" }\nneeds = [\"Bad_ID\"]\n", command), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCheck(tt.id, []byte(tt.data))

			if tt.wantErr && err == nil {
				t.Errorf("got %v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
