package budget

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

var ErrInvalid = errors.New("invalid budgets of shampoo")

type Settings struct {
	Repair int
}

func ParseSettings(data []byte) (Settings, error) {
	file := struct {
		Repair int `toml:"repair"`
	}{Repair: 3}
	meta, err := toml.Decode(string(data), &file)
	if err != nil {
		return Settings{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		return Settings{}, fmt.Errorf("%w: unknown fields %v", ErrInvalid, undecoded)
	}
	if file.Repair < 0 {
		return Settings{}, fmt.Errorf("%w: repair %d is negative", ErrInvalid, file.Repair)
	}
	return Settings{Repair: file.Repair}, nil
}
