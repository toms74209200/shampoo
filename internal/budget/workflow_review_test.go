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
		{name: "when Reviewing with questions within the limit then returns steady", stats: []Stat{MergedTask{Task: first, Questions: 1}, MergedTask{Task: second}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1}, want: Steady{}},
		{name: "when Reviewing with more tasks asking than the limit then returns a review naming them", stats: []Stat{MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 3}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1}, want: Review{Asked: []string{first, second}}},
		{name: "when Reviewing with more tasks repairing than the limit then returns a review naming them", stats: []Stat{MergedTask{Task: first, Repairs: 1}, MergedTask{Task: second, Repairs: 2, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1}, want: Review{Asked: []string{second}, Repaired: []string{first, second}}},
		{name: "when Reviewing with tasks asking before the window then returns steady", stats: []Stat{MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}, MergedTask{Task: third}, MergedTask{Task: third}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1}, want: Steady{}},
		{name: "when Reviewing with tasks asking before the latest review then returns steady", stats: []Stat{MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}, WorkflowReviewed{}, MergedTask{Task: third, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1}, want: Steady{}},
		{name: "when Reviewing with failed reviews up to the limit then returns a review", stats: []Stat{WorkflowReviewed{}, MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1}, want: Review{Asked: []string{first, second}}},
		{name: "when Reviewing with failed reviews beyond the limit then returns an escalation", stats: []Stat{WorkflowReviewed{}, MergedTask{Task: third}, WorkflowReviewed{}, MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1}, want: Escalate{Asked: []string{first, second}}},
		{name: "when Reviewing with a review followed by more tasks than the window then returns a review", stats: []Stat{WorkflowReviewed{}, MergedTask{Task: third}, WorkflowReviewed{}, MergedTask{Task: third}, MergedTask{Task: third}, MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 1}, want: Review{Asked: []string{first, second}}},
		{name: "when Reviewing with failed reviews before a human review then returns a review", stats: []Stat{WorkflowReviewed{}, HumanReviewed{}, MergedTask{Task: first, Questions: 1}, MergedTask{Task: second, Questions: 1}}, limits: settings.WorkflowReview{Tasks: 2, MaxQuestions: 1, MaxRepairs: 1, MaxFailedAttempts: 0}, want: Review{Asked: []string{first, second}}},
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
