package cli

const HelpText = `Usage: shampoo <command> [args]

  shampoo next                Print what to do next from state as JSON
  shampoo guard <transition>  Run edit/task guard by hand (no stdin)
  shampoo hook <transition>   Run guard for agent hooks (input on stdin)
  shampoo policies            Print policies (JSON via policies --json)
  shampoo ask                 Open an issue to ask a human (JSON stdin)
  shampoo init                Set up shampoo in a project (JSON stdin)
  shampoo review              Record a review of the tree (JSON stdin)
  shampoo help                Print what commands do (also -h, --help)
  shampoo version             Print build version (also -v, --version)
`
