package settings

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

var ErrInvalid = errors.New("invalid settings")

type Settings struct {
	RepairAttempts int
	Tasks          string
	Agent          string
}

func Parse(data []byte) (Settings, error) {
	var file struct {
		StateMachine struct {
			HarnessRepair struct {
				MaxFailedAttempts int `toml:"max-failed-attempts"`
			} `toml:"harness-repair"`
		} `toml:"state-machine"`
		Orchestrator struct {
			Tasks string `toml:"tasks"`
			Agent string `toml:"agent"`
		} `toml:"orchestrator"`
	}
	file.StateMachine.HarnessRepair.MaxFailedAttempts = 3
	file.Orchestrator.Agent = "claude -p"
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
	if file.Orchestrator.Agent == "" {
		return Settings{}, fmt.Errorf("%w: the agent of the orchestrator is empty", ErrInvalid)
	}
	return Settings{RepairAttempts: file.StateMachine.HarnessRepair.MaxFailedAttempts, Tasks: file.Orchestrator.Tasks, Agent: file.Orchestrator.Agent}, nil
}
