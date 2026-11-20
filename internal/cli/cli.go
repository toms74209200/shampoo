package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/toms74209200/shampoo/internal/guard"
	"github.com/toms74209200/shampoo/internal/orchestrator"
	"github.com/toms74209200/shampoo/internal/policy"
	"github.com/toms74209200/shampoo/internal/rule"
)

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var code int
	var output, message string
	exits := map[rule.Outcome]int{rule.TaskFailure: 1, rule.WorkflowFailure: 2, rule.Exceeded: 3, rule.Handed: 4}
	command, err := Parse(args)
	if err != nil {
		code, message = 2, err.Error()+"\n"
	}
	switch c := command.(type) {
	case Next:
		action, err := orchestrator.Next("")
		encoded, encodeErr := json.Marshal(action)
		if err == nil && encodeErr == nil {
			output = string(encoded) + "\n"
		} else {
			code, message = 2, fmt.Sprintf("shampoo next: %v\n", errors.Join(err, encodeErr))
		}
	case Policies:
		policies, err := orchestrator.Policies("")
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo policies: %v\n", err)
		}
		output = policy.Render(policies)
	case Ask:
		url, err := orchestrator.Ask("", stdin)
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo ask: %v\n", err)
		} else {
			output = url + "\n"
		}
	case Review:
		if err := orchestrator.Review("", stdin); err != nil {
			code, message = 2, fmt.Sprintf("shampoo review: %v\n", err)
		}
	case Guard:
		outcome, report, err := guard.Run("", c.Transition)
		code, message = exits[outcome], report
		if err != nil {
			message += fmt.Sprintf("shampoo guard: %v\n", err)
		}
	}
	if _, err := fmt.Fprint(stdout, output); err != nil {
		return 2
	}
	if _, err := fmt.Fprint(stderr, message); err != nil {
		return 2
	}
	return code
}
