package rule

import (
	"fmt"
	"slices"
)

type Transition string

const (
	Edit      Transition = "edit"
	Task      Transition = "task"
	PR        Transition = "pr"
	Milestone Transition = "milestone"
)

func ParseTransition(name string) (Transition, error) {
	transition := Transition(name)
	if !slices.Contains([]Transition{Edit, Task, PR, Milestone}, transition) {
		return "", fmt.Errorf("%w: transition %q", ErrInvalid, name)
	}
	return transition, nil
}
