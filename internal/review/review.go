package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

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
	for _, finding := range *review.Findings {
		marked := slices.ContainsFunc(policies, func(p policy.Policy) bool { return p.ID == finding.Policy && p.Review })
		if !marked || finding.Comment == "" {
			return nil, fmt.Errorf("%w: a finding names a policy marked for review and has a comment: %+v", ErrInvalid, finding)
		}
	}
	return *review.Findings, nil
}
