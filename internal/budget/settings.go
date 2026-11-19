package budget

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

var ErrInvalid = errors.New("invalid budget settings")

type Burn struct {
	Runs int
	Rate float64
}

type Settings struct {
	Window int
	Fast   Burn
	Slow   Burn
}

func ParseSettings(data []byte) (Settings, error) {
	file := struct {
		Window int `toml:"window"`
		Fast   struct {
			Runs int     `toml:"runs"`
			Burn float64 `toml:"burn"`
		} `toml:"fast"`
		Slow struct {
			Burn float64 `toml:"burn"`
		} `toml:"slow"`
	}{Window: 20}
	file.Fast.Runs, file.Fast.Burn, file.Slow.Burn = 5, 4, 1
	meta, err := toml.Decode(string(data), &file)
	if err != nil {
		return Settings{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		return Settings{}, fmt.Errorf("%w: unknown fields %v", ErrInvalid, undecoded)
	}
	if file.Window < 1 || file.Fast.Runs < 1 || file.Fast.Burn <= 0 || file.Slow.Burn <= 0 {
		return Settings{}, fmt.Errorf("%w: windows hold at least one run, and burn rates are positive", ErrInvalid)
	}
	return Settings{Window: file.Window, Fast: Burn{Runs: min(file.Fast.Runs, file.Window), Rate: file.Fast.Burn}, Slow: Burn{Runs: file.Window, Rate: file.Slow.Burn}}, nil
}
