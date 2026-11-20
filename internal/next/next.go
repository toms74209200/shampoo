package next

import "github.com/toms74209200/shampoo/internal/rule"

type State struct {
	Source  string
	Branch  string
	Default bool
	Verdict rule.Outcome
}

type Action struct {
	Action string `json:"action"`
	Source string `json:"source,omitempty"`
	Branch string `json:"branch,omitempty"`
}

func Decide(s State) Action {
	if s.Default {
		return Action{Action: "take-task", Source: s.Source}
	}
	switch s.Verdict {
	case rule.Pass:
		return Action{Action: "open-pr", Branch: s.Branch}
	case rule.Exceeded:
		return Action{Action: "repair-harness", Branch: s.Branch}
	case rule.Handed:
		return Action{Action: "hand-to-human", Branch: s.Branch}
	}
	return Action{Action: "continue", Branch: s.Branch}
}
