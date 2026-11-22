//go:build small

package usage

import (
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	tests := []struct {
		name    string
		read    []string
		lines   []string
		pending string
		beyond  int
		want    Tokens
	}{
		{name: "when Read with a message then returns its tokens by kind", lines: []string{`{"type":"assistant","message":{"id":"msg_a","usage":{"input_tokens":1,"cache_creation_input_tokens":2,"cache_read_input_tokens":3,"output_tokens":4}}}` + "\n"}, want: Tokens{Input: 1, CacheCreation: 2, CacheRead: 3, Output: 4}},
		{name: "when Read with a message written twice then returns its tokens once", lines: []string{`{"type":"assistant","message":{"id":"msg_a","usage":{"input_tokens":1,"cache_creation_input_tokens":2,"cache_read_input_tokens":3,"output_tokens":4}}}` + "\n", `{"type":"assistant","message":{"id":"msg_a","usage":{"input_tokens":1,"cache_creation_input_tokens":2,"cache_read_input_tokens":3,"output_tokens":4}}}` + "\n"}, want: Tokens{Input: 1, CacheCreation: 2, CacheRead: 3, Output: 4}},
		{name: "when Read with an offset at the end of a message then returns the tokens after it only", read: []string{`{"type":"assistant","message":{"id":"msg_a","usage":{"input_tokens":1,"cache_creation_input_tokens":2,"cache_read_input_tokens":3,"output_tokens":4}}}` + "\n"}, lines: []string{`{"type":"assistant","message":{"id":"msg_b","usage":{"input_tokens":10,"cache_creation_input_tokens":20,"cache_read_input_tokens":30,"output_tokens":40}}}` + "\n"}, want: Tokens{Input: 10, CacheCreation: 20, CacheRead: 30, Output: 40}},
		{name: "when Read with a line not yet ended then returns the tokens before it", lines: []string{`{"type":"assistant","message":{"id":"msg_a","usage":{"input_tokens":1,"cache_creation_input_tokens":2,"cache_read_input_tokens":3,"output_tokens":4}}}` + "\n"}, pending: `{"type":"assistant","message":{"id":"msg_b","usage":{"input_tokens":10,"cache_creation_input_tokens":20,"cache_read_input_tokens":30,"output_tokens":40}}}`, want: Tokens{Input: 1, CacheCreation: 2, CacheRead: 3, Output: 4}},
		{name: "when Read with an offset beyond the transcript then returns its tokens from the start", lines: []string{`{"type":"assistant","message":{"id":"msg_a","usage":{"input_tokens":1,"cache_creation_input_tokens":2,"cache_read_input_tokens":3,"output_tokens":4}}}` + "\n"}, beyond: 1000, want: Tokens{Input: 1, CacheCreation: 2, CacheRead: 3, Output: 4}},
		{name: "when Read with lines without usage then returns no tokens", lines: []string{`{"type":"user","message":{"role":"user"}}` + "\n", "not json\n"}, want: Tokens{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var transcript []byte
			for _, line := range tt.read {
				transcript = append(transcript, line...)
			}
			from := len(transcript) + tt.beyond
			for _, line := range tt.lines {
				transcript = append(transcript, line...)
			}
			wantEnd := len(transcript)
			transcript = append(transcript, tt.pending...)

			got, end := Read(transcript, from)

			if got != tt.want || end != wantEnd {
				t.Errorf("got %#v, %d, want %#v, %d", got, end, tt.want, wantEnd)
			}
		})
	}
}

func TestParseOffsets(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Offsets
		wantErr bool
	}{
		{name: "when ParseOffsets with nothing then returns no offsets", data: "", want: Offsets{}},
		{name: "when ParseOffsets with offsets then returns them", data: `{"/tmp/t.jsonl":12}`, want: Offsets{"/tmp/t.jsonl": 12}},
		{name: "when ParseOffsets with broken data then returns an error", data: "{", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOffsets([]byte(tt.data))

			if tt.wantErr != (err != nil) || !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, %v, want %#v", got, err, tt.want)
			}
		})
	}
}
