package cli

import (
	"fmt"
	"io"

	"github.com/toms74209200/shampoo/internal/guard"
	"github.com/toms74209200/shampoo/internal/rule"
)

func Run(args []string, stderr io.Writer) int {
	var code int
	var message string
	if len(args) != 2 || args[0] != "guard" {
		code, message = 2, "usage: shampoo guard <transition>\n"
	} else if transition, err := rule.ParseTransition(args[1]); err != nil {
		code, message = 2, fmt.Sprintf("shampoo guard: %v\n", err)
	} else if transition != rule.Edit {
		code, message = 2, fmt.Sprintf("shampoo guard: %s is not implemented yet\n", transition)
	} else {
		outcome, report, err := guard.Run("", transition)
		code, message = map[rule.Outcome]int{rule.TaskFailure: 1, rule.WorkflowFailure: 2}[outcome], report
		if err != nil {
			message += fmt.Sprintf("shampoo guard: %v\n", err)
		}
	}
	if _, err := fmt.Fprint(stderr, message); err != nil {
		return 2
	}
	return code
}
