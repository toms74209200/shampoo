package budget

import (
	"bytes"
	"encoding/json"
	"math"
	"slices"

	"github.com/toms74209200/shampoo/internal/settings"
	"github.com/toms74209200/shampoo/internal/usage"
)

type StatLine struct {
	Task      string        `json:"task,omitempty"`
	Questions int           `json:"questions,omitempty"`
	Repairs   int           `json:"repairs,omitempty"`
	Tokens    *usage.Tokens `json:"tokens,omitempty"`
	Changed   int           `json:"changed,omitempty"`
	Review    string        `json:"review,omitempty"`
	Issue     string        `json:"issue,omitempty"`
}

//sumtype:decl
type Stat interface{ stat() }

type MergedTask struct {
	Task      string
	Questions int
	Repairs   int
	Tokens    int
	Changed   int
}

type WorkflowReviewed struct{}

type HumanReviewed struct{ Issue string }

func (MergedTask) stat()       {}
func (WorkflowReviewed) stat() {}
func (HumanReviewed) stat()    {}

func ParseStats(data []byte) []Stat {
	var stats []Stat
	for _, raw := range bytes.Split(data, []byte("\n")) {
		var line StatLine
		if err := json.Unmarshal(raw, &line); err != nil {
			continue
		}
		tokens := 0
		if t := line.Tokens; t != nil {
			tokens = t.Input + t.CacheCreation + t.CacheRead + t.Output
		}
		parsed := map[string]Stat{
			"":         MergedTask{Task: line.Task, Questions: line.Questions, Repairs: line.Repairs, Tokens: tokens, Changed: line.Changed},
			"workflow": WorkflowReviewed{},
			"human":    HumanReviewed{Issue: line.Issue},
		}
		if stat, ok := parsed[line.Review]; ok {
			stats = append(stats, stat)
		}
	}
	return stats
}

//sumtype:decl
type Due interface{ due() }

type Steady struct{}

type Review struct {
	Asked    []string
	Repaired []string
	Costly   []string
}

type Escalate struct {
	Asked    []string
	Repaired []string
	Costly   []string
}

func (Steady) due()   {}
func (Review) due()   {}
func (Escalate) due() {}

func Reviewing(stats []Stat, limits settings.WorkflowReview) Due {
	merged := 0
	var reviews []int
	var humans []bool
	var since, costed []MergedTask
	for _, stat := range stats {
		switch stat := stat.(type) {
		case WorkflowReviewed:
			reviews = append(reviews, merged)
			humans = append(humans, false)
			since = nil
		case HumanReviewed:
			reviews = append(reviews, merged)
			humans = append(humans, true)
			since = nil
		case MergedTask:
			merged++
			since = append(since, stat)
			costed = append(costed, stat)
		}
	}
	costed = costed[max(0, len(costed)-limits.Tasks):]
	costed = slices.DeleteFunc(costed, func(task MergedTask) bool { return task.Tokens <= 0 || task.Changed <= 0 })
	var asked, repaired, costly []string
	recent := since[max(0, len(since)-limits.Tasks):]
	for _, stat := range recent {
		if stat.Questions > 0 {
			asked = append(asked, stat.Task)
		}
		if stat.Repairs > 0 {
			repaired = append(repaired, stat.Task)
		}
	}
	if len(costed) >= limits.MinCostTasks {
		for _, task := range Costly(costed, limits.CostDeviations) {
			if slices.ContainsFunc(recent, func(stat MergedTask) bool { return stat.Task == task }) {
				costly = append(costly, task)
			}
		}
	}
	if len(asked) <= limits.MaxQuestions && len(repaired) <= limits.MaxRepairs && len(costly) == 0 {
		return Steady{}
	}
	failed, after := 0, merged
	for i, at := range slices.Backward(reviews) {
		if humans[i] || after-at > limits.Tasks {
			break
		}
		failed++
		after = at
	}
	if failed > limits.MaxFailedAttempts {
		return Escalate{Asked: asked, Repaired: repaired, Costly: costly}
	}
	return Review{Asked: asked, Repaired: repaired, Costly: costly}
}

func Costly(tasks []MergedTask, limit float64) []string {
	if len(tasks) == 0 {
		return nil
	}
	var slopes, intercepts, residuals, deviations []float64
	for i, a := range tasks {
		for _, b := range tasks[i+1:] {
			if a.Changed != b.Changed {
				slopes = append(slopes, (math.Log(float64(b.Tokens))-math.Log(float64(a.Tokens)))/(math.Log(float64(b.Changed))-math.Log(float64(a.Changed))))
			}
		}
	}
	slope := 0.0
	if len(slopes) > 0 {
		slices.Sort(slopes)
		slope = (slopes[(len(slopes)-1)/2] + slopes[len(slopes)/2]) / 2
	}
	for _, task := range tasks {
		intercepts = append(intercepts, math.Log(float64(task.Tokens))-slope*math.Log(float64(task.Changed)))
	}
	slices.Sort(intercepts)
	intercept := (intercepts[(len(intercepts)-1)/2] + intercepts[len(intercepts)/2]) / 2
	for _, task := range tasks {
		residuals = append(residuals, math.Log(float64(task.Tokens))-intercept-slope*math.Log(float64(task.Changed)))
	}
	sorted := slices.Sorted(slices.Values(residuals))
	center := (sorted[(len(sorted)-1)/2] + sorted[len(sorted)/2]) / 2
	for _, residual := range residuals {
		deviations = append(deviations, math.Abs(residual-center))
	}
	slices.Sort(deviations)
	spread := 1.4826 * (deviations[(len(deviations)-1)/2] + deviations[len(deviations)/2]) / 2
	var costly []string
	for i, task := range tasks {
		if residuals[i]-center > limit*spread {
			costly = append(costly, task.Task)
		}
	}
	return costly
}
