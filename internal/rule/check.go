package rule

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/BurntSushi/toml"
)

type Check struct {
	ID      string
	Command string
	Size    Size
	Exit    Exit
	Needs   []string
}

func ParseCheck(id string, data []byte) (Check, error) {
	ids := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	if !ids.MatchString(id) {
		return Check{}, fmt.Errorf("%w: check id %q", ErrInvalid, id)
	}
	var file struct {
		Command string            `toml:"command"`
		Size    string            `toml:"size"`
		Exit    map[string]string `toml:"exit"`
		Needs   []string          `toml:"needs"`
	}
	meta, err := toml.Decode(string(data), &file)
	if err != nil {
		return Check{}, fmt.Errorf("%w: check %s: %v", ErrInvalid, id, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		return Check{}, fmt.Errorf("%w: check %s: unknown fields %v", ErrInvalid, id, undecoded)
	}
	if file.Command == "" {
		return Check{}, fmt.Errorf("%w: check %s: no command", ErrInvalid, id)
	}
	size := Size(file.Size)
	if !slices.Contains([]Size{Small, Medium, Large}, size) {
		return Check{}, fmt.Errorf("%w: check %s: size %q", ErrInvalid, id, file.Size)
	}
	exit, err := parseExit(file.Exit, []Outcome{TaskFailure, WorkflowFailure})
	if err != nil {
		return Check{}, fmt.Errorf("check %s: %w", id, err)
	}
	for _, need := range file.Needs {
		if !ids.MatchString(need) || need == id {
			return Check{}, fmt.Errorf("%w: check %s: needs %q", ErrInvalid, id, need)
		}
	}
	return Check{ID: id, Command: file.Command, Size: size, Exit: exit, Needs: file.Needs}, nil
}
