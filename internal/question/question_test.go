//go:build small

package question

import (
	"crypto/rand"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	branch := rand.Text()
	tests := []struct {
		name    string
		data    string
		want    Issue
		wantErr bool
	}{
		{name: "when Parse with a decision alternatives and trade-offs then returns the issue titled by the decision", data: `{"decision": "Merge by squash", "alternatives": "A merge commit", "tradeoffs": "A linear history"}`, want: Issue{Title: "Merge by squash", Body: "## Decision\n\nMerge by squash\n\n## Alternatives\n\nA merge commit\n\n## Trade-offs\n\nA linear history\n\n## Task\n\n" + branch + "\n"}},
		{name: "when Parse with a decision of several lines then returns the issue titled by its first line", data: `{"decision": "Merge by squash\nfor history", "alternatives": "A merge commit", "tradeoffs": "A linear history"}`, want: Issue{Title: "Merge by squash", Body: "## Decision\n\nMerge by squash\nfor history\n\n## Alternatives\n\nA merge commit\n\n## Trade-offs\n\nA linear history\n\n## Task\n\n" + branch + "\n"}},
		{name: "when Parse with empty alternatives then returns an error", data: `{"decision": "Merge by squash", "alternatives": " ", "tradeoffs": "A linear history"}`, wantErr: true},
		{name: "when Parse with no trade-offs then returns an error", data: `{"decision": "Merge by squash", "alternatives": "A merge commit"}`, wantErr: true},
		{name: "when Parse with an unknown field then returns an error", data: `{"decision": "Merge by squash", "alternatives": "A merge commit", "tradeoffs": "A linear history", "title": "x"}`, wantErr: true},
		{name: "when Parse with invalid JSON then returns an error", data: `{"decision"`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.data), branch)

			if tt.wantErr && err == nil {
				t.Errorf("got %#v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
