package rule

import "regexp"

type Paths struct {
	patterns []*regexp.Regexp
}

func (p Paths) Covers(changed []string) bool {
	if len(p.patterns) == 0 {
		return len(changed) > 0
	}
	for _, path := range changed {
		for _, pattern := range p.patterns {
			if pattern.MatchString(path) {
				return true
			}
		}
	}
	return false
}
