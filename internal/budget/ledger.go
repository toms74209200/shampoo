package budget

import (
	"bytes"
	"encoding/json"
	"slices"

	"github.com/toms74209200/shampoo/internal/rule"
)

type Line struct {
	Run        string       `json:"run"`
	Task       string       `json:"task"`
	Transition string       `json:"transition"`
	Rule       string       `json:"rule"`
	Check      string       `json:"check"`
	Outcome    rule.Outcome `json:"outcome"`
	Millis     int64        `json:"millis"`
}

type Summary struct {
	Task       string           `json:"task"`
	Attempts   map[string]int   `json:"attempts"`
	Violations map[string]int   `json:"violations"`
	Millis     map[string]int64 `json:"millis"`
	Repairs    int              `json:"repairs"`
}

func ParseLedger(data []byte) []Line {
	var lines []Line
	for _, raw := range bytes.Split(data, []byte("\n")) {
		var line Line
		if err := json.Unmarshal(raw, &line); err == nil {
			lines = append(lines, line)
		}
	}
	return lines
}

func Violations(lines []Line, ruleID string, runs int) int {
	violations, seen := 0, 0
	for _, line := range slices.Backward(lines) {
		if seen == runs {
			break
		}
		if line.Rule != ruleID {
			continue
		}
		seen++
		if line.Outcome == rule.TaskFailure || line.Outcome == rule.Budget {
			violations++
		}
	}
	return violations
}

func Repairs(lines []Line, task string) int {
	repairs := 0
	for _, line := range lines {
		if line.Task == task && line.Outcome == rule.Exceeded {
			repairs++
		}
	}
	return repairs
}

type Compaction struct {
	Kept      []Line
	Summaries []Summary
}

func Compact(lines []Line, open []string, window int) Compaction {
	latest := map[string]int{}
	keep := make([]bool, len(lines))
	for i, line := range slices.Backward(lines) {
		keep[i] = latest[line.Rule] < window || slices.Contains(open, line.Task)
		latest[line.Rule]++
	}
	var kept []Line
	var summaries []Summary
	index := map[string]int{}
	attempted := map[string]bool{}
	timed := map[[2]string]bool{}
	for i, line := range lines {
		merged := line.Task != "" && !slices.Contains(open, line.Task)
		if keep[i] {
			summarized := line
			if merged {
				summarized.Task = ""
			}
			kept = append(kept, summarized)
		}
		if !merged {
			continue
		}
		at, ok := index[line.Task]
		if !ok {
			at = len(summaries)
			index[line.Task] = at
			summaries = append(summaries, Summary{Task: line.Task, Attempts: map[string]int{}, Violations: map[string]int{}, Millis: map[string]int64{}})
		}
		if !attempted[line.Run] {
			attempted[line.Run] = true
			summaries[at].Attempts[line.Transition]++
		}
		if line.Outcome == rule.Exceeded {
			summaries[at].Repairs++
			continue
		}
		if !timed[[2]string{line.Run, line.Check}] {
			timed[[2]string{line.Run, line.Check}] = true
			summaries[at].Millis[line.Check] += line.Millis
		}
		if line.Outcome == rule.TaskFailure || line.Outcome == rule.Budget {
			summaries[at].Violations[line.Rule]++
		}
	}
	return Compaction{Kept: kept, Summaries: summaries}
}
