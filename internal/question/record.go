package question

import (
	"bytes"
	"encoding/json"
	"slices"
)

type Record struct {
	Issue string `json:"issue"`
	Kind  string `json:"kind"`
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
		kinds := map[string]*[]string{"question": &opened.Questions, "workflow": &opened.Workflows}
		if found, ok := kinds[record.Kind]; ok && !slices.Contains(*found, record.Issue) {
			*found = append(*found, record.Issue)
		}
	}
	return opened
}
