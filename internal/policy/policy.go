package policy

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/BurntSushi/toml"
)

var ErrInvalid = errors.New("invalid policy")

type Policy struct {
	ID      string
	Text    string
	Example string
	Review  bool
}

func Parse(id string, data []byte) (Policy, error) {
	if !regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(id) {
		return Policy{}, fmt.Errorf("%w: id %q", ErrInvalid, id)
	}
	var file struct {
		Text    string `toml:"text"`
		Example string `toml:"example"`
		Review  bool   `toml:"review"`
	}
	meta, err := toml.Decode(string(data), &file)
	if err != nil {
		return Policy{}, fmt.Errorf("%w: %s: %v", ErrInvalid, id, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		return Policy{}, fmt.Errorf("%w: %s: unknown fields %v", ErrInvalid, id, undecoded)
	}
	if file.Text == "" {
		return Policy{}, fmt.Errorf("%w: %s: no text", ErrInvalid, id)
	}
	return Policy{ID: id, Text: file.Text, Example: file.Example, Review: file.Review}, nil
}
