package rule

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

var ErrInvalid = errors.New("invalid rule")

type Rule interface{ rule() }

type Checked struct {
	ID          string
	Statement   string
	Command     string
	Transitions []Transition
	Paths       Paths
	Size        Size
	Exit        Exit
	Budget      float64
}

type Unchecked struct {
	ID        string
	Statement string
	Reason    string
}

func (Checked) rule()   {}
func (Unchecked) rule() {}

func Parse(id string, data []byte) (Rule, error) {
	if !regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(id) {
		return nil, fmt.Errorf("%w: id %q", ErrInvalid, id)
	}
	var file struct {
		Rule        string            `toml:"rule"`
		Check       string            `toml:"check"`
		Transitions []string          `toml:"transitions"`
		Paths       []string          `toml:"paths"`
		Size        string            `toml:"size"`
		Exit        map[string]string `toml:"exit"`
		Unchecked   string            `toml:"unchecked"`
		Budget      any               `toml:"budget"`
	}
	meta, err := toml.Decode(string(data), &file)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrInvalid, id, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		return nil, fmt.Errorf("%w: %s: unknown fields %v", ErrInvalid, id, undecoded)
	}
	if file.Rule == "" {
		return nil, fmt.Errorf("%w: %s: no rule", ErrInvalid, id)
	}
	if meta.IsDefined("unchecked") {
		if slices.ContainsFunc([]string{"check", "transitions", "paths", "size", "exit", "budget"}, func(key string) bool { return meta.IsDefined(key) }) || file.Unchecked == "" {
			return nil, fmt.Errorf("%w: %s: an unchecked rule has a reason and no check", ErrInvalid, id)
		}
		return Unchecked{ID: id, Statement: file.Rule, Reason: file.Unchecked}, nil
	}
	if file.Check == "" || len(file.Transitions) == 0 {
		return nil, fmt.Errorf("%w: %s: a rule has a check and its transitions, or a reason for having none", ErrInvalid, id)
	}
	checked := Checked{ID: id, Statement: file.Rule, Command: file.Check, Size: Size(file.Size)}
	for _, name := range file.Transitions {
		transition, err := ParseTransition(name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		checked.Transitions = append(checked.Transitions, transition)
	}
	if !slices.Contains([]Size{Small, Medium, Large}, checked.Size) {
		return nil, fmt.Errorf("%w: %s: size %q", ErrInvalid, id, file.Size)
	}
	for _, glob := range file.Paths {
		if glob == "" || strings.ContainsAny(glob, "[]!+") {
			return nil, fmt.Errorf("%w: %s: path %q", ErrInvalid, id, glob)
		}
		var expr strings.Builder
		for i := 0; i < len(glob); i++ {
			switch {
			case strings.HasPrefix(glob[i:], "**"):
				expr.WriteString(".*")
				i++
			case glob[i] == '*':
				expr.WriteString("[^/]*")
			case glob[i] == '?':
				expr.WriteString("[^/]?")
			default:
				expr.WriteString(regexp.QuoteMeta(glob[i : i+1]))
			}
		}
		checked.Paths.patterns = append(checked.Paths.patterns, regexp.MustCompile("^"+expr.String()+"$"))
	}
	exit, err := parseExit(file.Exit, []Outcome{TaskFailure, WorkflowFailure, Budget})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", id, err)
	}
	checked.Exit = exit
	switch budget := file.Budget.(type) {
	case nil:
		if checked.Exit.other == Budget || slices.Contains(slices.Collect(maps.Values(checked.Exit.codes)), Budget) {
			return nil, fmt.Errorf("%w: %s: an exit charging the budget needs a budget", ErrInvalid, id)
		}
	case int64:
		checked.Budget = float64(budget)
	case float64:
		checked.Budget = budget
	default:
		return nil, fmt.Errorf("%w: %s: budget %v is not a number", ErrInvalid, id, budget)
	}
	if checked.Budget < 0 || checked.Budget > 1 {
		return nil, fmt.Errorf("%w: %s: budget %v is not a share between 0 and 1", ErrInvalid, id, checked.Budget)
	}
	return checked, nil
}
