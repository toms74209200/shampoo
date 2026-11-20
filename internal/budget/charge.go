package budget

import (
	"fmt"

	"github.com/toms74209200/shampoo/internal/rule"
)

type Limits struct {
	Budget rule.Count
	Stop   rule.Count
}

type Charge struct {
	Report   string
	Exceeded bool
}

func ChargeRule(lines []Line, checked rule.Checked, transition rule.Transition, outcome rule.Outcome, exceptions *int, failures Limits) Charge {
	var charge Charge
	if checked.Exceptions != nil && exceptions != nil && *exceptions > checked.Exceptions.Limit {
		charge.Report += fmt.Sprintf("%s has %d exceptions in %s, over its limit of %d: fix the check rather than add exceptions\n", checked.ID, *exceptions, checked.Exceptions.File, checked.Exceptions.Limit)
		charge.Exceeded = true
	}
	if outcome == rule.WorkflowFailure {
		if used := Failures(lines, checked.ID, failures.Budget.Runs); used >= failures.Budget.Violations {
			charge.Report += fmt.Sprintf("%s burns its workflow-failure budget: %d workflow failures in the latest %d runs, of %d allowed\n", checked.ID, used, failures.Budget.Runs, failures.Budget.Violations)
		}
		if recent := Failures(lines, checked.ID, failures.Stop.Runs); recent >= failures.Stop.Violations {
			charge.Report += fmt.Sprintf("%s exceeds its workflow-failure budget: %d workflow failures in the latest %d runs: repair its check or its judge\n", checked.ID, recent, failures.Stop.Runs)
			charge.Exceeded = true
		}
	}
	if outcome == rule.TaskFailure || outcome == rule.Budget {
		allowed, stop := checked.Budget[transition], checked.Stop[transition]
		if used := Violations(lines, checked.ID, transition, allowed.Runs); allowed.Violations > 0 && used >= allowed.Violations {
			charge.Report += fmt.Sprintf("%s burns its budget at %s: %d violations in the latest %d runs, of %d allowed\n", checked.ID, transition, used, allowed.Runs, allowed.Violations)
		}
		if recent := Violations(lines, checked.ID, transition, stop.Runs); allowed.Violations == 0 || recent >= stop.Violations {
			charge.Report += fmt.Sprintf("%s exceeds its budget at %s: %d violations in the latest %d runs\n", checked.ID, transition, recent, stop.Runs)
			charge.Exceeded = true
		}
	}
	return charge
}

func ChargeTime(lines []Line, check rule.Check, seconds int, time Limits) Charge {
	var charge Charge
	limit := int64(seconds) * 1000
	if used := Overruns(lines, check.ID, limit, time.Budget.Runs); used >= time.Budget.Violations {
		charge.Report += fmt.Sprintf("check %s burns its time budget: %d runs over %ds of %s checks in the latest %d runs, of %d allowed\n", check.ID, used, seconds, check.Size, time.Budget.Runs, time.Budget.Violations)
	}
	if recent := Overruns(lines, check.ID, limit, time.Stop.Runs); recent >= time.Stop.Violations {
		charge.Report += fmt.Sprintf("check %s exceeds its time budget: %d runs over %ds of %s checks in the latest %d runs: reorganize the check, such as by splitting it or moving it to a later transition\n", check.ID, recent, seconds, check.Size, time.Stop.Runs)
		charge.Exceeded = true
	}
	return charge
}
