package policy

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

var ErrInvalid = errors.New("invalid policy")

type Policy struct {
	ID      string
	Text    string
	Example string
	Review  bool
}

func Parse(id string, data []byte) (Policy, error) {
	if !regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(id) {
		return Policy{}, fmt.Errorf("%w: id %q", ErrInvalid, id)
	}
	var file struct {
		Text    string `toml:"text"`
		Example string `toml:"example"`
		Review  bool   `toml:"review"`
	}
	meta, err := toml.Decode(string(data), &file)
	if err != nil {
		return Policy{}, fmt.Errorf("%w: %s: %v", ErrInvalid, id, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		return Policy{}, fmt.Errorf("%w: %s: unknown fields %v", ErrInvalid, id, undecoded)
	}
	if file.Text == "" {
		return Policy{}, fmt.Errorf("%w: %s: no text", ErrInvalid, id)
	}
	return Policy{ID: id, Text: file.Text, Example: file.Example, Review: file.Review}, nil
}

func Render(policies []Policy) string {
	var out strings.Builder
	if len(policies) > 0 {
		out.WriteString("# Policies of this project (.shampoo/policies/)\n")
	}
	for _, p := range policies {
		fmt.Fprintf(&out, "\n## %s\n\n%s\n", p.ID, strings.TrimSpace(p.Text))
		if p.Example != "" {
			fmt.Fprintf(&out, "\n%s\n", strings.TrimSpace(p.Example))
		}
	}
	if len(policies) > 0 {
		out.WriteString("\n")
	}
	out.WriteString(asking)
	return out.String()
}

const asking = `# Asking a human (shampoo)

When you cannot settle a decision, or a policy tells you to ask, decide, go on
with the work, and ask a human without waiting for the answer:

    echo '{"decision": "...", "alternatives": "...", "tradeoffs": "..."}' | shampoo ask

- decision: what you decided
- alternatives: how else it could be built
- tradeoffs: what each costs, such as ease of use

The answer comes later in the issue, as feedback for later tasks.
`
