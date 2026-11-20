//go:build small

package task

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    []Task
		wantErr bool
	}{
		{name: "when Parse with a task per line then returns them in order", data: "{\"id\":\"add-readme\",\"dod\":\"Add a README\"}\n\n{\"id\":\"fix-2\",\"dod\":\"Fix it\"}\n", want: []Task{{ID: "add-readme", DoD: "Add a README"}, {ID: "fix-2", DoD: "Fix it"}}},
		{name: "when Parse with no line then returns no task", data: "", want: nil},
		{name: "when Parse with a line that is not JSON then returns an error", data: "add-readme\n", wantErr: true},
		{name: "when Parse with an unknown field then returns an error", data: "{\"id\":\"a\",\"dod\":\"b\",\"title\":\"c\"}\n", wantErr: true},
		{name: "when Parse with an id outside lowercase words then returns an error", data: "{\"id\":\"Add README\",\"dod\":\"b\"}\n", wantErr: true},
		{name: "when Parse with an empty definition of done then returns an error", data: "{\"id\":\"a\",\"dod\":\"\"}\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.data))

			if tt.wantErr && err == nil {
				t.Errorf("got %v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
