package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/toms74209200/shampoo/internal/guard"
	"github.com/toms74209200/shampoo/internal/orchestrator"
	"github.com/toms74209200/shampoo/internal/rule"
)

func Run(args []string, stdout, stderr io.Writer) int {
	var code int
	var output, message string
	exits := map[rule.Outcome]int{rule.TaskFailure: 1, rule.WorkflowFailure: 2, rule.Exceeded: 3, rule.Handed: 4}
	if len(args) == 1 && args[0] == "next" {
		action, err := orchestrator.Next("")
		encoded, encodeErr := json.Marshal(action)
		if err == nil && encodeErr == nil {
			output = string(encoded) + "\n"
		} else {
			code, message = 2, fmt.Sprintf("shampoo next: %v\n", errors.Join(err, encodeErr))
		}
	} else if len(args) != 2 || args[0] != "guard" {
		code, message = 2, "usage: shampoo guard <transition> | shampoo next\n"
	} else if transition, err := rule.ParseTransition(args[1]); err != nil {
		code, message = 2, fmt.Sprintf("shampoo guard: %v\n", err)
	} else if transition != rule.Edit && transition != rule.Task {
		code, message = 2, fmt.Sprintf("shampoo guard: %s is checked by the repository's CI, not by shampoo\n", transition)
	} else {
		outcome, report, err := guard.Run("", transition)
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
