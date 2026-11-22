package usage

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Tokens struct {
	Input         int `json:"input"`
	CacheCreation int `json:"cache-creation"`
	CacheRead     int `json:"cache-read"`
	Output        int `json:"output"`
}

type entry struct {
	Message struct {
		ID    string `json:"id"`
		Usage *struct {
			Input         int `json:"input_tokens"`
			CacheCreation int `json:"cache_creation_input_tokens"`
			CacheRead     int `json:"cache_read_input_tokens"`
			Output        int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

func Read(transcript []byte, from int) (Tokens, int) {
	end := bytes.LastIndexByte(transcript, '\n') + 1
	if from > end {
		from = 0
	}
	messages := map[string]Tokens{}
	for _, raw := range bytes.Split(transcript[from:end], []byte("\n")) {
		var e entry
		if err := json.Unmarshal(raw, &e); err != nil || e.Message.Usage == nil || e.Message.ID == "" {
			continue
		}
		u := e.Message.Usage
		messages[e.Message.ID] = Tokens{Input: u.Input, CacheCreation: u.CacheCreation, CacheRead: u.CacheRead, Output: u.Output}
	}
	var total Tokens
	for _, t := range messages {
		total.Input += t.Input
		total.CacheCreation += t.CacheCreation
		total.CacheRead += t.CacheRead
		total.Output += t.Output
	}
	return total, end
}

type Offsets map[string]int

func ParseOffsets(data []byte) (Offsets, error) {
	offsets := Offsets{}
	if len(data) == 0 {
		return offsets, nil
	}
	if err := json.Unmarshal(data, &offsets); err != nil {
		return nil, fmt.Errorf("transcript offsets: %w", err)
	}
	return offsets, nil
}
