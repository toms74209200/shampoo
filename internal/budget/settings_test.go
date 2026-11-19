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
		{name: "when ParseSettings with no settings then returns the defaults", data: "", want: Settings{Window: 20, Fast: Burn{Runs: 5, Rate: 4}, Slow: Burn{Runs: 20, Rate: 1}}},
		{name: "when ParseSettings with every setting then returns them with the slow burn over the window", data: "window = 10\nfast = { runs = 2, burn = 3 }\nslow = { burn = 0.5 }\n", want: Settings{Window: 10, Fast: Burn{Runs: 2, Rate: 3}, Slow: Burn{Runs: 10, Rate: 0.5}}},
		{name: "when ParseSettings with invalid TOML then returns an error", data: "window = ", wantErr: true},
		{name: "when ParseSettings with an unknown field then returns an error", data: "length = 3\n", wantErr: true},
		{name: "when ParseSettings with a window of zero runs then returns an error", data: "window = 0\nfast = { runs = 1 }\n", wantErr: true},
		{name: "when ParseSettings with a fast burn over more runs than the window then returns the fast burn over the window", data: "window = 3\nfast = { runs = 4 }\n", want: Settings{Window: 3, Fast: Burn{Runs: 3, Rate: 4}, Slow: Burn{Runs: 3, Rate: 1}}},
		{name: "when ParseSettings with a fast burn over zero runs then returns an error", data: "fast = { runs = 0 }\n", wantErr: true},
		{name: "when ParseSettings with a fast burn rate of zero then returns an error", data: "fast = { burn = 0 }\n", wantErr: true},
		{name: "when ParseSettings with a negative slow burn rate then returns an error", data: "slow = { burn = -1 }\n", wantErr: true},
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
