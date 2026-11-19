package rule

import (
	"errors"
	"fmt"
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
	Check       string
	Judge       Judge
	Transitions []Transition
	Paths       Paths
	Budget      Count
	Stop        Count
}

type Count struct {
	Violations int
	Runs       int
}

type Unchecked struct {
	ID        string
	Statement string
	Reason    string
}

func (Checked) rule()   {}
func (Unchecked) rule() {}

type Judge interface{ judge() }

type ByExit struct {
	Exit Exit
}

type ByCommand struct {
	Command string
	Exit    Exit
}

func (ByExit) judge()    {}
func (ByCommand) judge() {}

func Parse(id string, data []byte) (Rule, error) {
	ids := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	if !ids.MatchString(id) {
		return nil, fmt.Errorf("%w: id %q", ErrInvalid, id)
	}
	var file struct {
		Rule        string   `toml:"rule"`
		Check       string   `toml:"check"`
		Transitions []string `toml:"transitions"`
		Paths       []string `toml:"paths"`
		Judge       *struct {
			Command string            `toml:"command"`
			Exit    map[string]string `toml:"exit"`
		} `toml:"judge"`
		Unchecked string `toml:"unchecked"`
		Budget    struct {
			Violations int `toml:"violations"`
			Runs       int `toml:"runs"`
		} `toml:"budget"`
		Stop struct {
			Violations int `toml:"violations"`
			Runs       int `toml:"runs"`
		} `toml:"stop"`
	}
	file.Budget.Violations, file.Budget.Runs, file.Stop.Violations, file.Stop.Runs = 2, 20, 2, 5
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
		if slices.ContainsFunc([]string{"check", "transitions", "paths", "judge", "budget", "stop"}, func(key string) bool { return meta.IsDefined(key) }) || file.Unchecked == "" {
			return nil, fmt.Errorf("%w: %s: an unchecked rule has a reason and no check", ErrInvalid, id)
		}
		return Unchecked{ID: id, Statement: file.Rule, Reason: file.Unchecked}, nil
	}
	if !ids.MatchString(file.Check) || len(file.Transitions) == 0 {
		return nil, fmt.Errorf("%w: %s: a rule has the id of its check and its transitions, or a reason for having none", ErrInvalid, id)
	}
	checked := Checked{ID: id, Statement: file.Rule, Check: file.Check}
	for _, name := range file.Transitions {
		transition, err := ParseTransition(name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		checked.Transitions = append(checked.Transitions, transition)
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
	if file.Judge == nil {
		return nil, fmt.Errorf("%w: %s: a rule has a judge", ErrInvalid, id)
	}
	exit, err := parseExit(file.Judge.Exit, []Outcome{TaskFailure, WorkflowFailure, Budget})
	if err != nil {
		return nil, fmt.Errorf("%s: judge: %w", id, err)
	}
	checked.Judge = ByExit{Exit: exit}
	if meta.IsDefined("judge", "command") {
		if file.Judge.Command == "" {
			return nil, fmt.Errorf("%w: %s: a judge command is not empty", ErrInvalid, id)
		}
		checked.Judge = ByCommand{Command: file.Judge.Command, Exit: exit}
	}
	if file.Budget.Runs < 1 || file.Budget.Violations < 0 || file.Budget.Violations > file.Budget.Runs || file.Stop.Runs < 1 || file.Stop.Violations < 1 || file.Stop.Violations > file.Stop.Runs {
		return nil, fmt.Errorf("%w: %s: a budget allows from none to all of at least one run, and a stop takes from one to all of at least one run", ErrInvalid, id)
	}
	checked.Budget = Count{Violations: file.Budget.Violations, Runs: file.Budget.Runs}
	checked.Stop = Count{Violations: file.Stop.Violations, Runs: file.Stop.Runs}
	return checked, nil
}
