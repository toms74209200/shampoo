package settings

import (
	"errors"
	"fmt"
	"slices"

	"github.com/BurntSushi/toml"
)

var ErrInvalid = errors.New("invalid settings")

type Settings struct {
	RepairAttempts int
	Tasks          string
	MaxWaitMinutes int
	Merge          string
	Requirements   map[string]any
}

func Parse(data []byte) (Settings, error) {
	var file struct {
		StateMachine struct {
			HarnessRepair struct {
				MaxFailedAttempts int `toml:"max-failed-attempts"`
			} `toml:"harness-repair"`
			Review struct {
				Requirements map[string]any `toml:"requirements"`
			} `toml:"review"`
		} `toml:"state-machine"`
		Orchestrator struct {
			Tasks          string `toml:"tasks"`
			MaxWaitMinutes int    `toml:"max-wait-minutes"`
			Merge          string `toml:"merge"`
		} `toml:"orchestrator"`
	}
	file.StateMachine.HarnessRepair.MaxFailedAttempts = 3
	file.Orchestrator.MaxWaitMinutes = 60
	file.Orchestrator.Merge = "merge"
	meta, err := toml.Decode(string(data), &file)
	if err != nil {
		return Settings{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		return Settings{}, fmt.Errorf("%w: unknown fields %v", ErrInvalid, undecoded)
	}
	if file.StateMachine.HarnessRepair.MaxFailedAttempts < 0 {
		return Settings{}, fmt.Errorf("%w: max-failed-attempts %d is negative", ErrInvalid, file.StateMachine.HarnessRepair.MaxFailedAttempts)
	}
	if file.Orchestrator.MaxWaitMinutes < 0 {
		return Settings{}, fmt.Errorf("%w: max-wait-minutes %d is negative", ErrInvalid, file.Orchestrator.MaxWaitMinutes)
	}
	if !slices.Contains([]string{"merge", "squash", "rebase"}, file.Orchestrator.Merge) {
		return Settings{}, fmt.Errorf("%w: merge %q is not merge, squash or rebase", ErrInvalid, file.Orchestrator.Merge)
	}
	return Settings{RepairAttempts: file.StateMachine.HarnessRepair.MaxFailedAttempts, Tasks: file.Orchestrator.Tasks, MaxWaitMinutes: file.Orchestrator.MaxWaitMinutes, Merge: file.Orchestrator.Merge, Requirements: file.StateMachine.Review.Requirements}, nil
}
