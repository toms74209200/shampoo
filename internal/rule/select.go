package rule

import "slices"

func Select(rules []Rule, transition Transition, changed []string) []Checked {
	var selected []Checked
	for _, r := range rules {
		if checked, ok := r.(Checked); ok && slices.Contains(checked.Transitions, transition) && checked.Paths.Covers(changed) {
			selected = append(selected, checked)
		}
	}
	return selected
}
