//go:build small

package budget

import (
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/toms74209200/shampoo/internal/settings"
)

func TestParseStats(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []Stat
	}{
		{name: "when ParseStats with summaries and reviews then returns them in order", data: "{\"task\":\"a\",\"attempts\":{},\"repairs\":1,\"questions\":2}\n{\"review\":\"workflow\"}\n{\"review\":\"human\",\"issue\":\"u\"}\n", want: []Stat{MergedTask{Task: "a", Questions: 2, Repairs: 1}, WorkflowReviewed{}, HumanReviewed{Issue: "u"}}},
		{name: "when ParseStats with a broken last line then returns the lines before it", data: "{\"task\":\"a\"}\n{\"task\":\"b", want: []Stat{MergedTask{Task: "a"}}},
		{name: "when ParseStats with an unknown review then returns the other lines", data: "{\"review\":\"other\"}\n{\"task\":\"a\"}\n", want: []Stat{MergedTask{Task: "a"}}},
		{name: "when ParseStats with tokens by kind then returns their total with the changed lines", data: "{\"task\":\"a\",\"tokens\":{\"input\":1,\"cache-creation\":2,\"cache-read\":3,\"output\":4},\"changed\":5}\n", want: []Stat{MergedTask{Task: "a", Tokens: 10, Changed: 5}}},
		{name: "when ParseStats with no data then returns no line", data: "", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStats([]byte(tt.data))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestReviewing(t *testing.T) {
	first, second, third := rand.Text(), rand.Text(), rand.Text()
	tests := []struct {
		name   string
		stats  []Stat
		limits settings.WorkflowReview
		want   Due
	}{
		{name: "when Reviewing with a costly task then returns a review naming it", stats: []Stat{MergedTask{Task: "t0", Tokens: 69570, Changed: 10}, MergedTask{Task: "t1", Tokens: 72000, Changed: 16}, MergedTask{Task: "t2", Tokens: 112178, Changed: 26}, MergedTask{Task: "t3", Tokens: 115256, Changed: 41}, MergedTask{Task: "t4", Tokens: 178729, Changed: 66}, MergedTask{Task: "t5", Tokens: 184445, Changed: 105}, MergedTask{Task: "t6", Tokens: 285153, Changed: 168}, MergedTask{Task: "t7", Tokens: 294673, Changed: 268}, MergedTask{Task: "t8", Tokens: 455671, Changed: 429}, MergedTask{Task: "t9", Tokens: 471792, Changed: 687}, MergedTask{Task: "t10", Tokens: 729657, Changed: 1100}, MergedTask{Task: "t11", Tokens: 754928, Changed: 1759}, MergedTask{Task: "x", Tokens: 1732051, Changed: 300}}, limits: settings.WorkflowReview{Tasks: 20, MaxQuestions: 10, MaxRepairs: 10, MaxFailedAttempts: 3, CostDeviations: 2, MinCostTasks: 10}, want: Review{Costly: []string{"x"}}},
		{name: "when Reviewing with a costly task among just enough to compare then returns a review naming it", stats: []Stat{MergedTask{Task: "t0", Tokens: 69570, Changed: 10}, MergedTask{Task: "t1", Tokens: 72000, Changed: 16}, MergedTask{Task: "t2", Tokens: 112178, Changed: 26}, MergedTask{Task: "t3", Tokens: 115256, Changed: 41}, MergedTask{Task: "t4", Tokens: 178729, Changed: 66}, MergedTask{Task: "t5", Tokens: 184445, Changed: 105}, MergedTask{Task: "t6", Tokens: 285153, Changed: 168}, MergedTask{Task: "t7", Tokens: 294673, Changed: 268}, MergedTask{Task: "t8", Tokens: 455671, Changed: 429}, MergedTask{Task: "x", Tokens: 1732051, Changed: 300}}, limits: settings.WorkflowReview{Tasks: 20, MaxQuestions: 10, MaxRepairs: 10, MaxFailedAttempts: 3, CostDeviations: 2, MinCostTasks: 10}, want: Review{Costly: []string{"x"}}},
		{name: "when Reviewing with a costly task among too few to compare then returns steady", stats: []Stat{MergedTask{Task: "t0", Tokens: 69570, Changed: 10}, MergedTask{Task: "t1", Tokens: 72000, Changed: 16}, MergedTask{Task: "t2", Tokens: 112178, Changed: 26}, MergedTask{Task: "t3", Tokens: 115256, Changed: 41}, MergedTask{Task: "t4", Tokens: 178729, Changed: 66}, MergedTask{Task: "t5", Tokens: 184445, Changed: 105}, MergedTask{Task: "t6", Tokens: 285153, Changed: 168}, MergedTask{Task: "t7", Tokens: 294673, Changed: 268}, MergedTask{Task: "x", Tokens: 1732051, Changed: 300}}, limits: settings.WorkflowReview{Tasks: 20, MaxQuestions: 10, MaxRepairs: 10, MaxFailedAttempts: 3, CostDeviations: 2, MinCostTasks: 10}, want: Steady{}},
		{name: "when Reviewing with a costly task before the latest review then returns steady", stats: []Stat{MergedTask{Task: "t0", Tokens: 69570, Changed: 10}, MergedTask{Task: "t1", Tokens: 72000, Changed: 16}, MergedTask{Task: "t2", Tokens: 112178, Changed: 26}, MergedTask{Task: "t3", Tokens: 115256, Changed: 41}, MergedTask{Task: "t4", Tokens: 178729, Changed: 66}, MergedTask{Task: "t5", Tokens: 184445, Changed: 105}, MergedTask{Task: "t6", Tokens: 285153, Changed: 168}, MergedTask{Task: "t7", Tokens: 294673, Changed: 268}, MergedTask{Task: "t8", Tokens: 455671, Changed: 429}, MergedTask{Task: "t9", Tokens: 471792, Changed: 687}, MergedTask{Task: "t10", Tokens: 729657, Changed: 1100}, MergedTask{Task: "t11", Tokens: 754928, Changed: 1759}, MergedTask{Task: "x", Tokens: 1732051, Changed: 300}, WorkflowReviewed{}, MergedTask{Task: "u0", Tokens: 69570, Changed: 10}}, limits: settings.WorkflowReview{Tasks: 20, MaxQuestions: 10, MaxRepairs: 10, MaxFailedAttempts: 3, CostDeviations: 2, MinCostTasks: 10}, want: Steady{}},
		{name: "when Reviewing with questions within the limit then returns steady", stats: []Stat{MergedTask{Task: first, Questions: 1}, MergedTask{Task: second}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1, CostDeviations: 2, MinCostTasks: 10}, want: Steady{}},
		{name: "when Reviewing with more tasks asking than the limit then returns a review naming them", stats: []Stat{MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 3}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1, CostDeviations: 2, MinCostTasks: 10}, want: Review{Asked: []string{first, second}}},
		{name: "when Reviewing with more tasks repairing than the limit then returns a review naming them", stats: []Stat{MergedTask{Task: first, Repairs: 1}, MergedTask{Task: second, Repairs: 2, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1, CostDeviations: 2, MinCostTasks: 10}, want: Review{Asked: []string{second}, Repaired: []string{first, second}}},
		{name: "when Reviewing with tasks asking before the window then returns steady", stats: []Stat{MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}, MergedTask{Task: third}, MergedTask{Task: third}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1, CostDeviations: 2, MinCostTasks: 10}, want: Steady{}},
		{name: "when Reviewing with tasks asking before the latest review then returns steady", stats: []Stat{MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}, WorkflowReviewed{}, MergedTask{Task: third, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1, CostDeviations: 2, MinCostTasks: 10}, want: Steady{}},
		{name: "when Reviewing with failed reviews up to the limit then returns a review", stats: []Stat{WorkflowReviewed{}, MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1, CostDeviations: 2, MinCostTasks: 10}, want: Review{Asked: []string{first, second}}},
		{name: "when Reviewing with failed reviews beyond the limit then returns an escalation", stats: []Stat{WorkflowReviewed{}, MergedTask{Task: third}, WorkflowReviewed{}, MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1, CostDeviations: 2, MinCostTasks: 10}, want: Escalate{Asked: []string{first, second}}},
		{name: "when Reviewing with a review followed by more tasks than the window then returns a review", stats: []Stat{WorkflowReviewed{}, MergedTask{Task: third}, WorkflowReviewed{}, MergedTask{Task: third}, MergedTask{Task: third}, MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1, CostDeviations: 2, MinCostTasks: 10}, want: Review{Asked: []string{first, second}}},
		{name: "when Reviewing with failed reviews before a human review then returns a review", stats: []Stat{WorkflowReviewed{}, HumanReviewed{}, MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 0, CostDeviations: 2, MinCostTasks: 10}, want: Review{Asked: []string{first, second}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reviewing(tt.stats, tt.limits)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCostly(t *testing.T) {
	tests := []struct {
		name  string
		tasks []MergedTask
		limit float64
		want  []string
	}{
		{name: "when Costly with no tasks then returns none", tasks: nil, limit: 2, want: nil},
		{name: "when Costly with a task far above what its size predicts then returns it", tasks: []MergedTask{{Task: "t0", Tokens: 69570, Changed: 10}, {Task: "t1", Tokens: 72000, Changed: 16}, {Task: "t2", Tokens: 112178, Changed: 26}, {Task: "t3", Tokens: 115256, Changed: 41}, {Task: "t4", Tokens: 178729, Changed: 66}, {Task: "t5", Tokens: 184445, Changed: 105}, {Task: "t6", Tokens: 285153, Changed: 168}, {Task: "t7", Tokens: 294673, Changed: 268}, {Task: "t8", Tokens: 455671, Changed: 429}, {Task: "t9", Tokens: 471792, Changed: 687}, {Task: "t10", Tokens: 729657, Changed: 1100}, {Task: "t11", Tokens: 754928, Changed: 1759}, {Task: "x", Tokens: 1732051, Changed: 300}}, limit: 2, want: []string{"x"}},
		{name: "when Costly with a task within the spread then returns none", tasks: []MergedTask{{Task: "t0", Tokens: 69570, Changed: 10}, {Task: "t1", Tokens: 72000, Changed: 16}, {Task: "t2", Tokens: 112178, Changed: 26}, {Task: "t3", Tokens: 115256, Changed: 41}, {Task: "t4", Tokens: 178729, Changed: 66}, {Task: "t5", Tokens: 184445, Changed: 105}, {Task: "t6", Tokens: 285153, Changed: 168}, {Task: "t7", Tokens: 294673, Changed: 268}, {Task: "t8", Tokens: 455671, Changed: 429}, {Task: "t9", Tokens: 471792, Changed: 687}, {Task: "t10", Tokens: 729657, Changed: 1100}, {Task: "t11", Tokens: 754928, Changed: 1759}, {Task: "x", Tokens: 415692, Changed: 300}}, limit: 2, want: nil},
		{name: "when Costly with a task far above a wider limit then returns none", tasks: []MergedTask{{Task: "t0", Tokens: 69570, Changed: 10}, {Task: "t1", Tokens: 72000, Changed: 16}, {Task: "t2", Tokens: 112178, Changed: 26}, {Task: "t3", Tokens: 115256, Changed: 41}, {Task: "t4", Tokens: 178729, Changed: 66}, {Task: "t5", Tokens: 184445, Changed: 105}, {Task: "t6", Tokens: 285153, Changed: 168}, {Task: "t7", Tokens: 294673, Changed: 268}, {Task: "t8", Tokens: 455671, Changed: 429}, {Task: "t9", Tokens: 471792, Changed: 687}, {Task: "t10", Tokens: 729657, Changed: 1100}, {Task: "t11", Tokens: 754928, Changed: 1759}, {Task: "x", Tokens: 1732051, Changed: 300}}, limit: 20, want: nil},
		{name: "when Costly with a task far below what its size predicts then returns none", tasks: []MergedTask{{Task: "t0", Tokens: 69570, Changed: 10}, {Task: "t1", Tokens: 72000, Changed: 16}, {Task: "t2", Tokens: 112178, Changed: 26}, {Task: "t3", Tokens: 115256, Changed: 41}, {Task: "t4", Tokens: 178729, Changed: 66}, {Task: "t5", Tokens: 184445, Changed: 105}, {Task: "t6", Tokens: 285153, Changed: 168}, {Task: "t7", Tokens: 294673, Changed: 268}, {Task: "t8", Tokens: 455671, Changed: 429}, {Task: "t9", Tokens: 471792, Changed: 687}, {Task: "t10", Tokens: 729657, Changed: 1100}, {Task: "t11", Tokens: 754928, Changed: 1759}, {Task: "x", Tokens: 69282, Changed: 300}}, limit: 2, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Costly(tt.tasks, tt.limit)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
