package question

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalid = errors.New("invalid question")

type Issue struct {
	Title string
	Body  string
}

func Parse(data []byte, branch string) (Issue, error) {
	var q struct {
		Decision     string `json:"decision"`
		Alternatives string `json:"alternatives"`
		Tradeoffs    string `json:"tradeoffs"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&q); err != nil {
		return Issue{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	decision, alternatives, tradeoffs := strings.TrimSpace(q.Decision), strings.TrimSpace(q.Alternatives), strings.TrimSpace(q.Tradeoffs)
	if decision == "" || alternatives == "" || tradeoffs == "" {
		return Issue{}, fmt.Errorf("%w: a question has a decision, alternatives and trade-offs, none empty", ErrInvalid)
	}
	title, _, _ := strings.Cut(decision, "\n")
	body := fmt.Sprintf("## Decision\n\n%s\n\n## Alternatives\n\n%s\n\n## Trade-offs\n\n%s\n\n## Task\n\n%s\n", decision, alternatives, tradeoffs, branch)
	return Issue{Title: title, Body: body}, nil
}
