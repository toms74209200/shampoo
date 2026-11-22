package budget

import "github.com/toms74209200/shampoo/internal/rule"

type Reviews struct {
	Attempts int
	Max      int
}

func Diverged(lines []Line, task string, limit Reviews) bool {
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
	return last >= 0 && counts[last] > 0 && (len(counts) > limit.Max || last > 0 && counts[last] >= counts[last-1] && charges > limit.Attempts)
}

func Reviewed(ledger []byte, run, task string, policies []string, limit Reviews) []Line {
	lines := []Line{{Run: run, Task: task, Transition: "review", Outcome: rule.Pass}}
	for _, policy := range policies {
		lines = append(lines, Line{Run: run, Task: task, Transition: "review", Policy: policy, Outcome: rule.Finding})
	}
	if Diverged(append(ParseLedger(ledger), lines...), task, limit) {
		lines = append(lines, Line{Run: run, Task: task, Transition: "review", Outcome: rule.Exceeded})
	}
	return lines
}

//sumtype:decl
type Convergence interface{ convergence() }

type Converging struct{}

type Diverging struct{}

type Exhausted struct{}

func (Converging) convergence() {}
func (Diverging) convergence()  {}
func (Exhausted) convergence()  {}

func Converge(lines []Line, task string, limit Reviews, repairAttempts int) Convergence {
	if !Diverged(lines, task, limit) {
		return Converging{}
	}
	if Repairs(lines, task) > repairAttempts {
		return Exhausted{}
	}
	return Diverging{}
}
