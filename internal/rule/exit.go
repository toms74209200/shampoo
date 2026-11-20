package rule

import (
	"fmt"
	"slices"
	"strconv"
)

type Outcome string

const (
	Pass            Outcome = "pass"
	TaskFailure     Outcome = "task-failure"
	WorkflowFailure Outcome = "workflow-failure"
	Budget          Outcome = "budget"
	Exceeded        Outcome = "budget-exceeded"
	Handed          Outcome = "repair-exceeded"
	Finding         Outcome = "finding"
	Question        Outcome = "question"
)

type Exit struct {
	codes map[int]Outcome
	other Outcome
}

func (e Exit) Outcome(code int) Outcome {
	if code == 0 {
		return Pass
	}
	if outcome, ok := e.codes[code]; ok {
		return outcome
	}
	return e.other
}

func parseExit(entries map[string]string, outcomes []Outcome) (Exit, error) {
	exit := Exit{codes: map[int]Outcome{}}
	for key, value := range entries {
		outcome := Outcome(value)
		if !slices.Contains(outcomes, outcome) {
			return Exit{}, fmt.Errorf("%w: exit outcome %q", ErrInvalid, value)
		}
		if key == "other" {
			exit.other = outcome
			continue
		}
		code, err := strconv.Atoi(key)
		if err != nil || code <= 0 {
			return Exit{}, fmt.Errorf("%w: exit code %q", ErrInvalid, key)
		}
		exit.codes[code] = outcome
	}
	if exit.other == "" {
		return Exit{}, fmt.Errorf("%w: exit has no other", ErrInvalid)
	}
	return exit, nil
}
