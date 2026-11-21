//go:build small

package question

import (
	"reflect"
	"testing"
)

func TestOpened(t *testing.T) {
	tests := []struct {
		name    string
		records string
		open    []string
		want    Open
	}{
		{name: "when Opened with recorded issues still open then returns them by kind", records: "{\"issue\":\"u1\",\"kind\":\"question\",\"task\":\"a\"}\n{\"issue\":\"u2\",\"kind\":\"workflow\",\"task\":\"main\"}\n", open: []string{"u1", "u2"}, want: Open{Questions: []string{"u1"}, Workflows: []string{"u2"}}},
		{name: "when Opened with a recorded issue closed then returns it not", records: "{\"issue\":\"u1\",\"kind\":\"question\"}\n{\"issue\":\"u2\",\"kind\":\"question\"}\n", open: []string{"u2"}, want: Open{Questions: []string{"u2"}}},
		{name: "when Opened with an open issue not recorded then returns it not", records: "{\"issue\":\"u1\",\"kind\":\"question\"}\n", open: []string{"u1", "u9"}, want: Open{Questions: []string{"u1"}}},
		{name: "when Opened with an issue recorded twice then returns it once", records: "{\"issue\":\"u1\",\"kind\":\"question\"}\n{\"issue\":\"u1\",\"kind\":\"question\"}\n", open: []string{"u1"}, want: Open{Questions: []string{"u1"}}},
		{name: "when Opened with an unknown kind or a broken line then returns the others", records: "{\"issue\":\"u1\",\"kind\":\"other\"}\n{\"issue\":\"u2\",\"ki", open: []string{"u1", "u2"}, want: Open{}},
		{name: "when Opened with no record then returns nothing", records: "", open: []string{"u1"}, want: Open{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Opened([]byte(tt.records), tt.open)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
