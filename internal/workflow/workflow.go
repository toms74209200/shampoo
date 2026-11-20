package workflow

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

var ErrInvalid = errors.New("invalid workflow")

type Workflow struct {
	Name           string
	PullRequest    bool
	branches       []filter
	branchesIgnore []filter
	paths          []filter
	pathsIgnore    []filter
}

type filter struct {
	expr    *regexp.Regexp
	negated bool
}

func Parse(path string, data []byte) (Workflow, error) {
	var file struct {
		Name string    `yaml:"name"`
		On   yaml.Node `yaml:"on"`
	}
	if err := yaml.Unmarshal(data, &file); err != nil {
		return Workflow{}, fmt.Errorf("%w: %s: %v", ErrInvalid, path, err)
	}
	parsed := Workflow{Name: file.Name}
	if parsed.Name == "" {
		parsed.Name = path
	}
	var events []*yaml.Node
	switch file.On.Kind {
	case yaml.ScalarNode:
		events = []*yaml.Node{&file.On}
	case yaml.SequenceNode:
		events = file.On.Content
	case yaml.MappingNode:
		for i := 0; i+1 < len(file.On.Content); i += 2 {
			events = append(events, file.On.Content[i])
		}
	}
	at := slices.IndexFunc(events, func(event *yaml.Node) bool { return event.Value == "pull_request" })
	if at < 0 {
		return parsed, nil
	}
	parsed.PullRequest = true
	if file.On.Kind != yaml.MappingNode {
		return parsed, nil
	}
	var filters struct {
		Branches       []string `yaml:"branches"`
		BranchesIgnore []string `yaml:"branches-ignore"`
		Paths          []string `yaml:"paths"`
		PathsIgnore    []string `yaml:"paths-ignore"`
	}
	if err := file.On.Content[2*at+1].Decode(&filters); err != nil {
		return Workflow{}, fmt.Errorf("%w: %s: pull_request: %v", ErrInvalid, path, err)
	}
	for _, list := range []struct {
		globs []string
		into  *[]filter
	}{
		{filters.Branches, &parsed.branches},
		{filters.BranchesIgnore, &parsed.branchesIgnore},
		{filters.Paths, &parsed.paths},
		{filters.PathsIgnore, &parsed.pathsIgnore},
	} {
		for _, glob := range list.globs {
			negated := strings.HasPrefix(glob, "!")
			glob = strings.TrimPrefix(glob, "!")
			var expr strings.Builder
			for i := 0; i < len(glob); i++ {
				switch {
				case strings.HasPrefix(glob[i:], "**"):
					expr.WriteString(".*")
					i++
				case glob[i] == '*':
					expr.WriteString("[^/]*")
				case strings.ContainsRune("?+[]", rune(glob[i])):
					expr.WriteByte(glob[i])
				default:
					expr.WriteString(regexp.QuoteMeta(glob[i : i+1]))
				}
			}
			compiled, err := regexp.Compile("^" + expr.String() + "$")
			if err != nil || glob == "" {
				return Workflow{}, fmt.Errorf("%w: %s: pattern %q", ErrInvalid, path, glob)
			}
			*list.into = append(*list.into, filter{expr: compiled, negated: negated})
		}
	}
	return parsed, nil
}

func (w Workflow) Expects(base string, changed []string) bool {
	matches := func(filters []filter, name string) bool {
		matched := false
		for _, f := range filters {
			if f.expr.MatchString(name) {
				matched = !f.negated
			}
		}
		return matched
	}
	switch {
	case !w.PullRequest:
		return false
	case len(w.branches) > 0 && !matches(w.branches, base):
		return false
	case len(w.branchesIgnore) > 0 && matches(w.branchesIgnore, base):
		return false
	case len(w.paths) > 0:
		return slices.ContainsFunc(changed, func(path string) bool { return matches(w.paths, path) })
	case len(w.pathsIgnore) > 0:
		return slices.ContainsFunc(changed, func(path string) bool { return !matches(w.pathsIgnore, path) })
	}
	return true
}

type Source struct {
	Path string
	Data []byte
}

func Expected(sources []Source, base string, changed []string) ([]string, error) {
	var names []string
	for _, source := range sources {
		parsed, err := Parse(source.Path, source.Data)
		if err != nil {
			return nil, err
		}
		if parsed.Expects(base, changed) {
			names = append(names, parsed.Name)
		}
	}
	return names, nil
}
