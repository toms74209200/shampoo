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
		{name: "when ParseSettings with no settings then returns the defaults", data: "", want: Settings{RepairAttempts: 3}},
		{name: "when ParseSettings with failed attempts of harness repair then returns them", data: "[state-machine.harness-repair]\nmax-failed-attempts = 1\n", want: Settings{RepairAttempts: 1}},
		{name: "when ParseSettings with invalid TOML then returns an error", data: "[state-machine", wantErr: true},
		{name: "when ParseSettings with a state not settable yet then returns an error", data: "[state-machine.editing]\nmax-failed-attempts = 3\n", wantErr: true},
		{name: "when ParseSettings with negative failed attempts then returns an error", data: "[state-machine.harness-repair]\nmax-failed-attempts = -1\n", wantErr: true},
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
