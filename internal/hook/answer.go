package hook

import (
	"github.com/toms74209200/shampoo/internal/cli"
	"github.com/toms74209200/shampoo/internal/rule"
)

type Answer struct {
	Code              int          `json:"code"`
	Outcome           rule.Outcome `json:"outcome"`
	Report            string       `json:"report"`
	Decision          string       `json:"decision,omitempty"`
	Reason            string       `json:"reason,omitempty"`
	FollowupMessage   string       `json:"followup_message,omitempty"`
	AdditionalContext string       `json:"additional_context,omitempty"`
	SystemMessage     string       `json:"systemMessage,omitempty"`
}

func Guard(outcome rule.Outcome, report string) Answer {
	answer := Answer{Code: cli.Exits[outcome], Outcome: outcome, Report: report}
	switch outcome {
	case rule.TaskFailure, rule.Exceeded:
		answer.Decision, answer.Reason, answer.FollowupMessage, answer.AdditionalContext = "block", report, report, report
	case rule.WorkflowFailure, rule.Handed:
		answer.SystemMessage = report
	}
	return answer
}

type Context struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
	AdditionalContext string `json:"additional_context"`
}

func Policies(text string) Context {
	var context Context
	context.HookSpecificOutput.HookEventName = "SessionStart"
	context.HookSpecificOutput.AdditionalContext = text
	context.AdditionalContext = text
	return context
}
