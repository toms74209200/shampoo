package cli

import (
	"errors"
	"fmt"

	"github.com/toms74209200/shampoo/internal/rule"
)

var ErrUsage = errors.New("usage: shampoo guard <transition> | shampoo next | shampoo review | shampoo policies | shampoo ask")

//sumtype:decl
type Command interface{ command() }

type Guard struct{ Transition rule.Transition }

type Next struct{}

type Review struct{}

type Policies struct{}

type Ask struct{}

func (Guard) command()    {}
func (Next) command()     {}
func (Review) command()   {}
func (Policies) command() {}
func (Ask) command()      {}

func Parse(args []string) (Command, error) {
	commands := map[string]Command{"next": Next{}, "review": Review{}, "policies": Policies{}, "ask": Ask{}}
	if len(args) == 1 {
		if command, ok := commands[args[0]]; ok {
			return command, nil
		}
	}
	if len(args) != 2 || args[0] != "guard" {
		return nil, ErrUsage
	}
	transition, err := rule.ParseTransition(args[1])
	if err != nil {
		return nil, fmt.Errorf("shampoo guard: %w", err)
	}
	if transition != rule.Edit && transition != rule.Task {
		return nil, fmt.Errorf("shampoo guard: %s is checked by the repository's CI, not by shampoo", transition)
	}
	return Guard{Transition: transition}, nil
}
