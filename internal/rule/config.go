package rule

import (
	"fmt"
	"slices"
)

type Config struct {
	rules  []Rule
	checks map[string]Check
}

func Link(rules []Rule, checks []Check) (Config, error) {
	config := Config{rules: rules, checks: map[string]Check{}}
	for _, check := range checks {
		if _, ok := config.checks[check.ID]; ok {
			return Config{}, fmt.Errorf("%w: check %s is defined twice", ErrInvalid, check.ID)
		}
		config.checks[check.ID] = check
	}
	for _, r := range rules {
		if checked, ok := r.(Checked); ok {
			if _, ok := config.checks[checked.Check]; !ok {
				return Config{}, fmt.Errorf("%w: %s: no check %s", ErrInvalid, checked.ID, checked.Check)
			}
		}
	}
	visiting, done := map[string]bool{}, map[string]bool{}
	var visit func(id string) error
	visit = func(id string) error {
		if done[id] {
			return nil
		}
		if visiting[id] {
			return fmt.Errorf("%w: check %s needs itself", ErrInvalid, id)
		}
		visiting[id] = true
		for _, need := range config.checks[id].Needs {
			if _, ok := config.checks[need]; !ok {
				return fmt.Errorf("%w: check %s: no check %s to need", ErrInvalid, id, need)
			}
			if err := visit(need); err != nil {
				return err
			}
		}
		done[id] = true
		return nil
	}
	for _, check := range checks {
		if err := visit(check.ID); err != nil {
			return Config{}, err
		}
	}
	return config, nil
}

func (c Config) Plan(transition Transition, changed []string) ([]Checked, []Check) {
	var selected []Checked
	var order []Check
	var visit func(id string)
	visit = func(id string) {
		if slices.ContainsFunc(order, func(check Check) bool { return check.ID == id }) {
			return
		}
		for _, need := range c.checks[id].Needs {
			visit(need)
		}
		order = append(order, c.checks[id])
	}
	for _, r := range c.rules {
		if checked, ok := r.(Checked); ok && slices.Contains(checked.Transitions, transition) && checked.Paths.Covers(changed) {
			selected = append(selected, checked)
			visit(checked.Check)
		}
	}
	return selected, order
}
