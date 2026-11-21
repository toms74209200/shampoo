package budget

import "github.com/toms74209200/shampoo/internal/rule"

func Diverged(lines []Line, task string, attempts int) bool {
	var counts []int
	runs := map[string]int{}
	for _, line := range lines {
		if line.Task != task || line.Transition != "review" {
			continue
		}
		at, seen := runs[line.Run]
		if !seen {
			at = len(counts)
			runs[line.Run] = at
			counts = append(counts, 0)
		}
		if line.Outcome == rule.Finding {
			counts[at]++
		}
	}
	charges := 0
	for i := 1; i < len(counts); i++ {
		if counts[i] >= counts[i-1] {
			charges++
		}
	}
	last := len(counts) - 1
	return last > 0 && counts[last] > 0 && counts[last] >= counts[last-1] && charges > attempts
}

func Reviewed(ledger []byte, run, task string, policies []string, attempts int) []Line {
	lines := []Line{{Run: run, Task: task, Transition: "review", Outcome: rule.Pass}}
	for _, policy := range policies {
		lines = append(lines, Line{Run: run, Task: task, Transition: "review", Policy: policy, Outcome: rule.Finding})
	}
	if Diverged(append(ParseLedger(ledger), lines...), task, attempts) {
		lines = append(lines, Line{Run: run, Task: task, Transition: "review", Outcome: rule.Exceeded})
	}
	return lines
}
