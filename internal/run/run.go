package run

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/toms74209200/shampoo/internal/cli"
	"github.com/toms74209200/shampoo/internal/guard"
	"github.com/toms74209200/shampoo/internal/hook"
	"github.com/toms74209200/shampoo/internal/orchestrator"
	"github.com/toms74209200/shampoo/internal/policy"
	"github.com/toms74209200/shampoo/internal/rule"
)

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var code int
	var output, message string
	command, err := cli.Parse(args)
	if err != nil {
		code, message = 2, err.Error()+"\n"
	}
	switch c := command.(type) {
	case cli.Next:
		action, err := orchestrator.Next("")
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo next: %v\n", err)
			break
		}
		encoded, err := json.Marshal(action)
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo next: %v\n", err)
			break
		}
		output = string(encoded) + "\n"
	case cli.Policies:
		policies, err := orchestrator.Policies("")
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo policies: %v\n", err)
		}
		output = policy.Render(policies)
	case cli.Ask:
		url, err := orchestrator.Ask("", stdin)
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo ask: %v\n", err)
		} else {
			output = url + "\n"
		}
	case cli.Init:
		report, notice, err := orchestrator.Init("", stdin)
		message = notice
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo init: %v\n", err)
			break
		}
		encoded, err := json.Marshal(report)
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo init: %v\n", err)
			break
		}
		output = string(encoded) + "\n"
	case cli.Review:
		if err := orchestrator.Review("", stdin); err != nil {
			code, message = 2, fmt.Sprintf("shampoo review: %v\n", err)
		}
	case cli.GuardHook:
		input, err := io.ReadAll(stdin)
		if err == nil {
			err = guard.Usage("", hook.Stop(c.Transition, input))
		}
		outcome, report := rule.WorkflowFailure, ""
		if err == nil {
			outcome, report, err = guard.Run("", c.Transition)
		}
		if err != nil {
			report += fmt.Sprintf("shampoo guard: %v\n", err)
		}
		encoded, err := json.Marshal(hook.Guard(outcome, report))
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo guard: %v\n", err)
			break
		}
		output = string(encoded) + "\n"
	case cli.PoliciesHook:
		policies, err := orchestrator.Policies("")
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo policies: %v\n", err)
			break
		}
		encoded, err := json.Marshal(hook.Policies(policy.Render(policies)))
		if err != nil {
			code, message = 2, fmt.Sprintf("shampoo policies: %v\n", err)
			break
		}
		output = string(encoded) + "\n"
	case cli.Guard:
		outcome, report, err := guard.Run("", c.Transition)
		code, message = cli.Exits[outcome], report
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
