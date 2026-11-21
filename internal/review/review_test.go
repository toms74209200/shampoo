//go:build small

package review

import (
	"reflect"
	"strings"
	"testing"

	"github.com/toms74209200/shampoo/internal/policy"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		policies    []policy.Policy
		data        string
		want        []Finding
		wantErr     bool
		wantMessage string
	}{
		{name: "when Parse with a finding on a policy for review then returns the finding", policies: []policy.Policy{{ID: "functions", Review: true}, {ID: "layout"}}, data: `{"findings": [{"policy": "functions", "comment": "a wrapper"}]}`, want: []Finding{{Policy: "functions", Comment: "a wrapper"}}},
		{name: "when Parse with no finding then returns an empty review", policies: []policy.Policy{{ID: "functions", Review: true}}, data: `{"findings": []}`, want: []Finding{}},
		{name: "when Parse with a finding on a policy not for review then returns an error naming it and the policies for review", policies: []policy.Policy{{ID: "functions", Review: true}, {ID: "tests", Review: true}, {ID: "layout"}}, data: `{"findings": [{"policy": "layout", "comment": "a wrapper"}]}`, wantErr: true, wantMessage: `policy "layout" is not marked for review; the policies marked for review are functions, tests`},
		{name: "when Parse with a finding on an unknown policy then returns an error naming it and the policies for review", policies: []policy.Policy{{ID: "functions", Review: true}}, data: `{"findings": [{"policy": "comments", "comment": "a wrapper"}]}`, wantErr: true, wantMessage: `policy "comments" does not exist; the policies marked for review are functions`},
		{name: "when Parse with a finding and no policy for review then returns an error saying so", policies: []policy.Policy{{ID: "layout"}}, data: `{"findings": [{"policy": "layout", "comment": "a wrapper"}]}`, wantErr: true, wantMessage: `no policy is marked for review`},
		{name: "when Parse with a finding without a comment then returns an error naming its policy", policies: []policy.Policy{{ID: "functions", Review: true}}, data: `{"findings": [{"policy": "functions"}]}`, wantErr: true, wantMessage: `the finding on policy "functions" has no comment`},
		{name: "when Parse with no findings field then returns an error", policies: []policy.Policy{{ID: "functions", Review: true}}, data: `{}`, wantErr: true, wantMessage: `no findings`},
		{name: "when Parse with invalid JSON then returns an error", policies: []policy.Policy{{ID: "functions", Review: true}}, data: `{`, wantErr: true, wantMessage: `invalid review`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.data), tt.policies)

			if tt.wantErr && (err == nil || !strings.Contains(err.Error(), tt.wantMessage)) {
				t.Errorf("got %#v, %v, want an error with %q", got, err, tt.wantMessage)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
