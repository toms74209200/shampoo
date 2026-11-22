package hook

import (
	"encoding/json"

	"github.com/toms74209200/shampoo/internal/rule"
)

//sumtype:decl
type Transcript interface{ transcript() }

type Named struct{ Path string }

type Unnamed struct{}

func (Named) transcript()   {}
func (Unnamed) transcript() {}

func Stop(transition rule.Transition, input []byte) Transcript {
	var hook struct {
		TranscriptPath string `json:"transcript_path"`
	}
	if transition != rule.Task || json.Unmarshal(input, &hook) != nil || hook.TranscriptPath == "" {
		return Unnamed{}
	}
	return Named{Path: hook.TranscriptPath}
}
