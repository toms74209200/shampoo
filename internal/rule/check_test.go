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
			name: "when ParseCheck with a command its size and its needs then returns the check",
			id:   id,
			data: fmt.Sprintf("command = %q\nsize = \"medium\"\nneeds = [%q]\n", command, need),
			want: Check{ID: id, Command: command, Size: Size("medium"), Needs: []string{need}},
		},
		{
			name: "when ParseCheck with no needs then returns the check needing nothing",
			id:   id,
			data: fmt.Sprintf("command = %q\nsize = \"large\"\n", command),
			want: Check{ID: id, Command: command, Size: Size("large")},
		},
		{name: "when ParseCheck with an id outside lowercase words then returns an error", id: "Bad_ID", data: fmt.Sprintf("command = %q\nsize = \"small\"\n", command), wantErr: true},
		{name: "when ParseCheck with invalid TOML then returns an error", id: id, data: "command = ", wantErr: true},
		{name: "when ParseCheck with an exit mapping then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nexit = { other = \"task-failure\" }\n", command), wantErr: true},
		{name: "when ParseCheck with no command then returns an error", id: id, data: "size = \"small\"\n", wantErr: true},
		{name: "when ParseCheck with an unknown size then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"huge\"\n", command), wantErr: true},
		{name: "when ParseCheck with a need of itself then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nneeds = [%q]\n", command, id), wantErr: true},
		{name: "when ParseCheck with a need outside lowercase words then returns an error", id: id, data: fmt.Sprintf("command = %q\nsize = \"small\"\nneeds = [\"Bad_ID\"]\n", command), wantErr: true},
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
