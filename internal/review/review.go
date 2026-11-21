package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/toms74209200/shampoo/internal/policy"
)

var ErrInvalid = errors.New("invalid review")

type Finding struct {
	Policy  string `json:"policy"`
	Comment string `json:"comment"`
}

func Parse(data []byte, policies []policy.Policy) ([]Finding, error) {
	var review struct {
		Findings *[]Finding `json:"findings"`
	}
	if err := json.Unmarshal(data, &review); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if review.Findings == nil {
		return nil, fmt.Errorf("%w: no findings; an empty list records a review that found nothing", ErrInvalid)
	}
	var marked []string
	for _, p := range policies {
		if p.Review {
			marked = append(marked, p.ID)
		}
	}
	known := "the policies marked for review are " + strings.Join(marked, ", ")
	if len(marked) == 0 {
		known = "no policy is marked for review"
	}
	for _, finding := range *review.Findings {
		exists := slices.ContainsFunc(policies, func(p policy.Policy) bool { return p.ID == finding.Policy })
		switch {
		case !exists:
			return nil, fmt.Errorf("%w: policy %q does not exist; %s", ErrInvalid, finding.Policy, known)
		case !slices.Contains(marked, finding.Policy):
			return nil, fmt.Errorf("%w: policy %q is not marked for review; %s", ErrInvalid, finding.Policy, known)
		case finding.Comment == "":
			return nil, fmt.Errorf("%w: the finding on policy %q has no comment", ErrInvalid, finding.Policy)
		}
	}
	return *review.Findings, nil
}
