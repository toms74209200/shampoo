package cli

import (
	"errors"
	"fmt"

	"github.com/toms74209200/shampoo/internal/rule"
)

var Exits = map[rule.Outcome]int{rule.TaskFailure: 1, rule.WorkflowFailure: 2, rule.Exceeded: 3, rule.Handed: 4}

var ErrUsage = errors.New("usage: shampoo guard <transition> | shampoo hook <transition> | shampoo next | shampoo review | shampoo policies [--json] | shampoo ask | shampoo init")

//sumtype:decl
type Command interface{ command() }

type Guard struct{ Transition rule.Transition }

type GuardHook struct{ Transition rule.Transition }

type Next struct{}

type Review struct{}

type Policies struct{}

type PoliciesHook struct{}

type Ask struct{}

type Init struct{}

func (Guard) command()        {}
func (GuardHook) command()    {}
func (PoliciesHook) command() {}
func (Next) command()         {}
func (Review) command()       {}
func (Policies) command()     {}
func (Ask) command()          {}
func (Init) command()         {}

func Parse(args []string) (Command, error) {
	commands := map[string]Command{"next": Next{}, "review": Review{}, "policies": Policies{}, "ask": Ask{}, "init": Init{}}
	if len(args) == 1 {
		if command, ok := commands[args[0]]; ok {
			return command, nil
		}
	}
	if len(args) == 2 && args[0] == "policies" && args[1] == "--json" {
		return PoliciesHook{}, nil
	}
	if len(args) != 2 || (args[0] != "guard" && args[0] != "hook") {
		return nil, ErrUsage
	}
	transition, err := rule.ParseTransition(args[1])
	if err != nil {
		return nil, fmt.Errorf("shampoo %s: %w", args[0], err)
	}
	if transition != rule.Edit && transition != rule.Task {
		return nil, fmt.Errorf("shampoo %s: %s is checked by the repository's CI, not by shampoo", args[0], transition)
	}
	if args[0] == "hook" {
		return GuardHook{Transition: transition}, nil
	}
	return Guard{Transition: transition}, nil
}
