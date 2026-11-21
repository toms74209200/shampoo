package budget

import (
	"bytes"
	"encoding/json"
	"slices"

	"github.com/toms74209200/shampoo/internal/settings"
)

type StatLine struct {
	Task      string `json:"task,omitempty"`
	Questions int    `json:"questions,omitempty"`
	Repairs   int    `json:"repairs,omitempty"`
	Review    string `json:"review,omitempty"`
	Issue     string `json:"issue,omitempty"`
}

//sumtype:decl
type Stat interface{ stat() }

type MergedTask struct {
	Task      string
	Questions int
	Repairs   int
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
		parsed := map[string]Stat{
			"":         MergedTask{Task: line.Task, Questions: line.Questions, Repairs: line.Repairs},
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
}

type Escalate struct {
	Asked    []string
	Repaired []string
}

func (Steady) due()   {}
func (Review) due()   {}
func (Escalate) due() {}

func Reviewing(stats []Stat, limits settings.WorkflowReview) Due {
	merged := 0
	var reviews []int
	var humans []bool
	var since []MergedTask
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
		}
	}
	var asked, repaired []string
	for _, stat := range since[max(0, len(since)-limits.Tasks):] {
		if stat.Questions > 0 {
			asked = append(asked, stat.Task)
		}
		if stat.Repairs > 0 {
			repaired = append(repaired, stat.Task)
		}
	}
	if len(asked) <= limits.MaxQuestions && len(repaired) <= limits.MaxRepairs {
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
		return Escalate{Asked: asked, Repaired: repaired}
	}
	return Review{Asked: asked, Repaired: repaired}
}
