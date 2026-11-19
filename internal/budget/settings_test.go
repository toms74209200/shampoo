//go:build small

package budget

import (
	"reflect"
	"testing"
)

func TestParseSettings(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Settings
		wantErr bool
	}{
		{name: "when ParseSettings with no settings then returns the defaults", data: "", want: Settings{Repair: 3}},
		{name: "when ParseSettings with a repair budget then returns it", data: "repair = 1\n", want: Settings{Repair: 1}},
		{name: "when ParseSettings with invalid TOML then returns an error", data: "repair = ", wantErr: true},
		{name: "when ParseSettings with an unknown field then returns an error", data: "window = 3\n", wantErr: true},
		{name: "when ParseSettings with a negative repair budget then returns an error", data: "repair = -1\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSettings([]byte(tt.data))

			if tt.wantErr && err == nil {
				t.Errorf("got %v, want an error", got)
			}
			if !tt.wantErr && (err != nil || !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
