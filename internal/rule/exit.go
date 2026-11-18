package rule

type Outcome string

const (
	Pass            Outcome = "pass"
	TaskFailure     Outcome = "task-failure"
	WorkflowFailure Outcome = "workflow-failure"
	StructureBudget Outcome = "structure-budget"
)

type Exit struct {
	codes map[int]Outcome
	other Outcome
}

func (e Exit) Outcome(code int) Outcome {
	if code == 0 {
		return Pass
	}
	if outcome, ok := e.codes[code]; ok {
		return outcome
	}
	return e.other
}
