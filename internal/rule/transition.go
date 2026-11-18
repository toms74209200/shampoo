package rule

type Transition string

const (
	Edit      Transition = "edit"
	Task      Transition = "task"
	PR        Transition = "pr"
	Milestone Transition = "milestone"
)
