//go:build small

package review

import (
	"reflect"
	"testing"

	"github.com/toms74209200/shampoo/internal/policy"
)

func TestParse(t *testing.T) {
	policies := []policy.Policy{{ID: "functions", Review: true}, {ID: "layout"}}
	tests := []struct {
		name    string
		data    string
		want    []Finding
		wantErr bool
	}{
		{name: "when Parse with a finding on a policy for review then returns the finding", data: `{"findings": [{"policy": "functions", "comment": "a wrapper"}]}`, want: []Finding{{Policy: "functions", Comment: "a wrapper"}}},
		{name: "when Parse with no finding then returns an empty review", data: `{"findings": []}`, want: []Finding{}},
		{name: "when Parse with a finding on a policy not for review then returns an error", data: `{"findings": [{"policy": "layout", "comment": "a wrapper"}]}`, wantErr: true},
		{name: "when Parse with a finding on an unknown policy then returns an error", data: `{"findings": [{"policy": "tests", "comment": "a wrapper"}]}`, wantErr: true},
		{name: "when Parse with a finding without a comment then returns an error", data: `{"findings": [{"policy": "functions"}]}`, wantErr: true},
		{name: "when Parse with no findings field then returns an error", data: `{}`, wantErr: true},
		{name: "when Parse with invalid JSON then returns an error", data: `{`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.data), policies)

			if tt.wantErr && err == nil {
				t.Errorf("got %#v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
