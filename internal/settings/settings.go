package settings

import (
	"errors"
	"fmt"
	"slices"

	"github.com/BurntSushi/toml"

	"github.com/toms74209200/shampoo/internal/rule"
)

var ErrInvalid = errors.New("invalid settings")

type Settings struct {
	RepairAttempts int
	ReviewAttempts int
	MaxReviews     int
	Tasks          string
	MaxWaitMinutes int
	Merge          string
	Requirements   map[string]any
	MaxSeconds     map[string]int
	TimeBudget     rule.Count
	TimeStop       rule.Count
	FailureBudget  rule.Count
	FailureStop    rule.Count
	MaxOpen        int
	WorkflowReview WorkflowReview
}

type WorkflowReview struct {
	Tasks             int     `toml:"tasks"`
	MaxQuestions      int     `toml:"max-questions"`
	MaxRepairs        int     `toml:"max-repairs"`
	MaxFailedAttempts int     `toml:"max-failed-attempts"`
	CostDeviations    float64 `toml:"cost-deviations"`
	MinCostTasks      int     `toml:"min-cost-tasks"`
}

func Parse(data []byte) (Settings, error) {
	var file struct {
		StateMachine struct {
			HarnessRepair struct {
				MaxFailedAttempts int `toml:"max-failed-attempts"`
			} `toml:"harness-repair"`
			Review struct {
				Requirements      map[string]any `toml:"requirements"`
				MaxFailedAttempts int            `toml:"max-failed-attempts"`
				MaxReviews        int            `toml:"max-reviews"`
			} `toml:"review"`
			WorkflowReview WorkflowReview `toml:"workflow-review"`
		} `toml:"state-machine"`
		Orchestrator struct {
			Tasks          string `toml:"tasks"`
			MaxWaitMinutes int    `toml:"max-wait-minutes"`
			Merge          string `toml:"merge"`
		} `toml:"orchestrator"`
		TestSizes map[string]struct {
			MaxSeconds int `toml:"max-seconds"`
		} `toml:"test-sizes"`
		TimeBudget struct {
			Budget rule.Count `toml:"budget"`
			Stop   rule.Count `toml:"stop"`
		} `toml:"time-budget"`
		WorkflowFailure struct {
			Budget rule.Count `toml:"budget"`
			Stop   rule.Count `toml:"stop"`
		} `toml:"workflow-failure"`
		HumanReview struct {
			MaxOpen int `toml:"max-open"`
		} `toml:"human-review"`
	}
	file.StateMachine.HarnessRepair.MaxFailedAttempts = 3
	file.StateMachine.Review.MaxFailedAttempts = 2
	file.StateMachine.Review.MaxReviews = 5
	file.Orchestrator.MaxWaitMinutes = 60
	file.Orchestrator.Merge = "merge"
	file.TimeBudget.Budget = rule.Count{Violations: 2, Runs: 20}
	file.TimeBudget.Stop = rule.Count{Violations: 2, Runs: 5}
	file.WorkflowFailure.Budget = rule.Count{Violations: 2, Runs: 20}
	file.WorkflowFailure.Stop = rule.Count{Violations: 2, Runs: 5}
	file.HumanReview.MaxOpen = 3
	file.StateMachine.WorkflowReview = WorkflowReview{Tasks: 20, MaxQuestions: 10, MaxRepairs: 10, MaxFailedAttempts: 3, CostDeviations: 2, MinCostTasks: 10}
	meta, err := toml.Decode(string(data), &file)
	if err != nil {
		return Settings{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		return Settings{}, fmt.Errorf("%w: unknown fields %v", ErrInvalid, undecoded)
	}
	if file.StateMachine.Review.MaxFailedAttempts < 0 {
		return Settings{}, fmt.Errorf("%w: max-failed-attempts %d of the review is negative", ErrInvalid, file.StateMachine.Review.MaxFailedAttempts)
	}
	if file.StateMachine.Review.MaxReviews < 1 {
		return Settings{}, fmt.Errorf("%w: max-reviews %d of the review is less than one", ErrInvalid, file.StateMachine.Review.MaxReviews)
	}
	if file.StateMachine.HarnessRepair.MaxFailedAttempts < 0 {
		return Settings{}, fmt.Errorf("%w: max-failed-attempts %d is negative", ErrInvalid, file.StateMachine.HarnessRepair.MaxFailedAttempts)
	}
	if file.Orchestrator.MaxWaitMinutes < 0 {
		return Settings{}, fmt.Errorf("%w: max-wait-minutes %d is negative", ErrInvalid, file.Orchestrator.MaxWaitMinutes)
	}
	maxSeconds := map[string]int{"small": 60, "medium": 300, "large": 900}
	for size, limit := range file.TestSizes {
		if _, ok := maxSeconds[size]; !ok {
			return Settings{}, fmt.Errorf("%w: test size %q is not small, medium or large", ErrInvalid, size)
		}
		if limit.MaxSeconds <= 0 {
			return Settings{}, fmt.Errorf("%w: max-seconds %d of %s is not positive", ErrInvalid, limit.MaxSeconds, size)
		}
		maxSeconds[size] = limit.MaxSeconds
	}
	counts := map[string]rule.Count{
		"budget of the time budget":       file.TimeBudget.Budget,
		"stop of the time budget":         file.TimeBudget.Stop,
		"budget of the workflow failures": file.WorkflowFailure.Budget,
		"stop of the workflow failures":   file.WorkflowFailure.Stop,
	}
	for name, count := range counts {
		if count.Runs < 1 || count.Violations < 1 || count.Violations > count.Runs {
			return Settings{}, fmt.Errorf("%w: %s has violations, from 1 to all, in runs, at least one", ErrInvalid, name)
		}
	}
	if file.HumanReview.MaxOpen < 0 {
		return Settings{}, fmt.Errorf("%w: human-review has max-open of at least zero", ErrInvalid)
	}
	if review := file.StateMachine.WorkflowReview; review.Tasks < 1 || review.MaxQuestions < 0 || review.MaxRepairs < 0 || review.MaxFailedAttempts < 0 || review.CostDeviations < 0 || review.MinCostTasks < 2 {
		return Settings{}, fmt.Errorf("%w: workflow-review has tasks of at least one, max-questions, max-repairs, max-failed-attempts and cost-deviations of at least zero, and min-cost-tasks of at least two", ErrInvalid)
	}
	if !slices.Contains([]string{"merge", "squash", "rebase"}, file.Orchestrator.Merge) {
		return Settings{}, fmt.Errorf("%w: merge %q is not merge, squash or rebase", ErrInvalid, file.Orchestrator.Merge)
	}
	return Settings{RepairAttempts: file.StateMachine.HarnessRepair.MaxFailedAttempts, ReviewAttempts: file.StateMachine.Review.MaxFailedAttempts, MaxReviews: file.StateMachine.Review.MaxReviews, Tasks: file.Orchestrator.Tasks, MaxWaitMinutes: file.Orchestrator.MaxWaitMinutes, Merge: file.Orchestrator.Merge, Requirements: file.StateMachine.Review.Requirements, MaxSeconds: maxSeconds, TimeBudget: file.TimeBudget.Budget, TimeStop: file.TimeBudget.Stop, FailureBudget: file.WorkflowFailure.Budget, FailureStop: file.WorkflowFailure.Stop, MaxOpen: file.HumanReview.MaxOpen, WorkflowReview: file.StateMachine.WorkflowReview}, nil
}
