package question

import (
	"bytes"
	"encoding/json"
	"slices"
)

type Kind string

const (
	Asked     Kind = "question"
	Escalated Kind = "workflow"
)

type Record struct {
	Issue string `json:"issue"`
	Kind  Kind   `json:"kind"`
	Task  string `json:"task"`
}

type Open struct {
	Questions []string
	Workflows []string
}

func Opened(records []byte, open []string) Open {
	var opened Open
	for _, raw := range bytes.Split(records, []byte("\n")) {
		var record Record
		if err := json.Unmarshal(raw, &record); err != nil || !slices.Contains(open, record.Issue) {
			continue
		}
		kinds := map[Kind]*[]string{Asked: &opened.Questions, Escalated: &opened.Workflows}
		if found, ok := kinds[record.Kind]; ok && !slices.Contains(*found, record.Issue) {
			*found = append(*found, record.Issue)
		}
	}
	return opened
}
