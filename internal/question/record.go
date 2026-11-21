package question

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func Records(data []byte) []Record {
	var records []Record
	for _, raw := range bytes.Split(data, []byte("\n")) {
		var record Record
		if err := json.Unmarshal(raw, &record); err == nil && slices.Contains([]Kind{Asked, Escalated}, record.Kind) {
			records = append(records, record)
		}
	}
	return records
}

func Query(records []Record) string {
	if len(records) == 0 {
		return "query { __typename }"
	}
	query := "query {"
	for i, record := range records {
		query += fmt.Sprintf(" i%d: resource(url: %q) { ... on Issue { state } }", i, record.Issue)
	}
	return query + " }"
}

func Opened(records []Record, response []byte) (Open, error) {
	var answer struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(response, &answer); err != nil {
		return Open{}, fmt.Errorf("%w: the states of the issues: %v", ErrInvalid, err)
	}
	var opened Open
	for i, record := range records {
		var issue struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal(answer.Data[fmt.Sprintf("i%d", i)], &issue); err != nil || issue.State != "OPEN" {
			continue
		}
		kinds := map[Kind]*[]string{Asked: &opened.Questions, Escalated: &opened.Workflows}
		if found := kinds[record.Kind]; !slices.Contains(*found, record.Issue) {
			*found = append(*found, record.Issue)
		}
	}
	return opened, nil
}
