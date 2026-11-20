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
	MaxSeconds     map[string]int
	TimeBudget     Count
	TimeStop       Count
}

type Count struct {
	Violations int `toml:"violations"`
	Runs       int `toml:"runs"`
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
		TestSizes map[string]struct {
			MaxSeconds int `toml:"max-seconds"`
		} `toml:"test-sizes"`
		TimeBudget struct {
			Budget Count `toml:"budget"`
			Stop   Count `toml:"stop"`
		} `toml:"time-budget"`
	}
	file.StateMachine.HarnessRepair.MaxFailedAttempts = 3
	file.Orchestrator.MaxWaitMinutes = 60
	file.Orchestrator.Merge = "merge"
	file.TimeBudget.Budget = Count{Violations: 2, Runs: 20}
	file.TimeBudget.Stop = Count{Violations: 2, Runs: 5}
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
	maxSeconds := map[string]int{"small": 60, "medium": 300, "large": 900}
	for size, limit := range file.TestSizes {
		if _, ok := maxSeconds[size]; !ok {
			return Settings{}, fmt.Errorf("%w: test size %q is not small, medium or large", ErrInvalid, size)
		}
		if limit.MaxSeconds <= 0 {
			return Settings{}, fmt.Errorf("%w: max-seconds %d of %s is not positive", ErrInvalid, limit.MaxSeconds, size)
		}
		maxSeconds[size] = limit.MaxSeconds
	}
	for name, count := range map[string]Count{"budget": file.TimeBudget.Budget, "stop": file.TimeBudget.Stop} {
		if count.Runs < 1 || count.Violations < 1 || count.Violations > count.Runs {
			return Settings{}, fmt.Errorf("%w: %s of the time budget has violations, from 1 to all, in runs, at least one", ErrInvalid, name)
		}
	}
	if !slices.Contains([]string{"merge", "squash", "rebase"}, file.Orchestrator.Merge) {
		return Settings{}, fmt.Errorf("%w: merge %q is not merge, squash or rebase", ErrInvalid, file.Orchestrator.Merge)
	}
	return Settings{RepairAttempts: file.StateMachine.HarnessRepair.MaxFailedAttempts, Tasks: file.Orchestrator.Tasks, MaxWaitMinutes: file.Orchestrator.MaxWaitMinutes, Merge: file.Orchestrator.Merge, Requirements: file.StateMachine.Review.Requirements, MaxSeconds: maxSeconds, TimeBudget: file.TimeBudget.Budget, TimeStop: file.TimeBudget.Stop}, nil
}
