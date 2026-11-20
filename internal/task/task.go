package task

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

var ErrInvalid = errors.New("invalid task")

type Task struct {
	ID  string
	DoD string
}

func Parse(data []byte) ([]Task, error) {
	ids := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	var tasks []Task
	for _, raw := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(raw)) == 0 {
			continue
		}
		var line struct {
			ID  string `json:"id"`
			DoD string `json:"dod"`
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&line); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
		}
		if !ids.MatchString(line.ID) || line.DoD == "" {
			return nil, fmt.Errorf("%w: a task has an id of lowercase words and a definition of done: %s", ErrInvalid, raw)
		}
		tasks = append(tasks, Task{ID: line.ID, DoD: line.DoD})
	}
	return tasks, nil
}
