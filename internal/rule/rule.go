package rule

import (
	"errors"
	"fmt"
	"path/filepath"
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
	Budget      map[Transition]Count
	Stop        map[Transition]Count
	Exceptions  *Exceptions
}

type Exceptions struct {
	File  string
	Limit int
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
		Unchecked  string         `toml:"unchecked"`
		Budget     map[string]any `toml:"budget"`
		Stop       map[string]any `toml:"stop"`
		Exceptions *struct {
			File  string `toml:"file"`
			Limit *int   `toml:"limit"`
		} `toml:"exceptions"`
	}
	meta, err := toml.Decode(string(data), &file)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrInvalid, id, err)
	}
	undecoded := slices.DeleteFunc(meta.Undecoded(), func(key toml.Key) bool { return key[0] == "budget" || key[0] == "stop" })
	if len(undecoded) > 0 {
		return nil, fmt.Errorf("%w: %s: unknown fields %v", ErrInvalid, id, undecoded)
	}
	if (meta.IsDefined("budget") && file.Budget == nil) || (meta.IsDefined("stop") && file.Stop == nil) {
		return nil, fmt.Errorf("%w: %s: a budget and a stop are counts", ErrInvalid, id)
	}
	if file.Rule == "" {
		return nil, fmt.Errorf("%w: %s: no rule", ErrInvalid, id)
	}
	if meta.IsDefined("unchecked") {
		if slices.ContainsFunc([]string{"check", "transitions", "paths", "judge", "budget", "stop", "exceptions"}, func(key string) bool { return meta.IsDefined(key) }) || file.Unchecked == "" {
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
	checked.Budget, err = parseCounts(file.Budget, checked.Transitions, 0, Count{Violations: 10, Runs: 20}, Count{Violations: 2, Runs: 20})
	if err != nil {
		return nil, fmt.Errorf("%s: budget: %w", id, err)
	}
	checked.Stop, err = parseCounts(file.Stop, checked.Transitions, 1, Count{Violations: 8, Runs: 10}, Count{Violations: 2, Runs: 5})
	if err != nil {
		return nil, fmt.Errorf("%s: stop: %w", id, err)
	}
	if file.Exceptions != nil {
		limit := 3
		if file.Exceptions.Limit != nil {
			limit = *file.Exceptions.Limit
		}
		path := file.Exceptions.File
		if path == "" || filepath.IsAbs(path) || !filepath.IsLocal(path) || limit < 0 {
			return nil, fmt.Errorf("%w: %s: exceptions have a file inside the repository and a limit of at least zero", ErrInvalid, id)
		}
		checked.Exceptions = &Exceptions{File: path, Limit: limit}
	}
	return checked, nil
}

func parseCounts(value map[string]any, transitions []Transition, least int, atEdit, elsewhere Count) (map[Transition]Count, error) {
	counts := map[Transition]Count{}
	for _, transition := range transitions {
		counts[transition] = elsewhere
	}
	if _, ok := counts[Edit]; ok {
		counts[Edit] = atEdit
	}
	_, hasViolations := value["violations"]
	_, hasRuns := value["runs"]
	if hasViolations || hasRuns {
		count, err := parseCount(value, least)
		if err != nil {
			return nil, err
		}
		for transition := range counts {
			counts[transition] = count
		}
		return counts, nil
	}
	for name, entry := range value {
		if _, ok := counts[Transition(name)]; !ok {
			return nil, fmt.Errorf("%w: %q is not a transition the rule guards", ErrInvalid, name)
		}
		table, ok := entry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: %s is not a count", ErrInvalid, name)
		}
		count, err := parseCount(table, least)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		counts[Transition(name)] = count
	}
	return counts, nil
}

func parseCount(table map[string]any, least int) (Count, error) {
	violations, hasViolations := table["violations"].(int64)
	runs, hasRuns := table["runs"].(int64)
	if !hasViolations || !hasRuns || len(table) != 2 || runs < 1 || violations < int64(least) || violations > runs {
		return Count{}, fmt.Errorf("%w: a count has violations, from %d to all, in runs, at least one", ErrInvalid, least)
	}
	return Count{Violations: int(violations), Runs: int(runs)}, nil
}
