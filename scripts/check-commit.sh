#!/bin/sh
# Checks a commit message against the commit rules (docs/development.md).
#
# Usage:
#   check-commit.sh --msg-file <file>  check a message file (commit-msg hook)
#   check-commit.sh --rev <rev>        check the message of an existing commit (CI)
#
# Exit codes: 0 pass, 1 rule violation (task failure), 2 checker error (workflow failure).
#
# Rules:
#   1. Subject is ":<gitmoji>: <Verb> <what>", using the emoji and verbs listed below
#   2. Subject is at most 72 characters and does not end with a period
#   3. A body, if any, is separated from the subject by a blank line

set -u

EMOJIS='
:sparkles:           new feature
:bug:                bug fix
:recycle:            refactoring
:art:                code structure or format
:white_check_mark:   tests
:memo:               documentation
:hammer:             development scripts and tooling
:construction_worker: CI
:rotating_light:     linter config or lint fixes
:see_no_evil:        .gitignore
:wrench:             configuration
:heavy_plus_sign:    add a dependency
:heavy_minus_sign:   remove a dependency
:arrow_up:           upgrade a dependency
:fire:               remove code or files
:truck:              move or rename
:zap:                performance
:pencil2:            typo
:rewind:             revert
'

VERBS='Add Change Update Bump bump Fix Remove Rename Move Refactor Clean Replace Apply Revert'

workflow_failure() {
	echo "check-commit: $1" >&2
	exit 2
}

case "${1:-}" in
--msg-file)
	[ $# -eq 2 ] || workflow_failure "usage: check-commit.sh --msg-file <file>"
	[ -r "$2" ] || workflow_failure "cannot read message file: $2"
	msg=$(git stripspace --strip-comments <"$2") || workflow_failure "git stripspace failed"
	;;
--rev)
	[ $# -eq 2 ] || workflow_failure "usage: check-commit.sh --rev <rev>"
	msg=$(git log -1 --format=%B "$2") || workflow_failure "cannot read commit: $2"
	;;
*)
	workflow_failure "usage: check-commit.sh --msg-file <file> | --rev <rev>"
	;;
esac

violations=''
violate() {
	violations="${violations}
- $1"
}

subject=$(printf '%s\n' "$msg" | sed -n '1p')
second=$(printf '%s\n' "$msg" | sed -n '2p')

emoji=$(printf '%s' "$subject" | sed -n 's/^\(:[a-z0-9_]*:\) .*/\1/p')
verb=$(printf '%s' "$subject" | sed -n 's/^:[a-z0-9_]*: \([^ ]*\) .*/\1/p')

if [ -z "$emoji" ] || [ -z "$verb" ]; then
	violate "subject must be \":<gitmoji>: <Verb> <what>\", got: \"$subject\""
else
	if ! printf '%s\n' "$EMOJIS" | grep -q "^$emoji "; then
		violate "emoji $emoji is not allowed; use one of:
$(printf '%s' "$EMOJIS" | sed '/^$/d; s/^/    /')"
	fi
	case " $VERBS " in
	*" $verb "*) ;;
	*) violate "verb \"$verb\" is not allowed; start with an imperative verb from: $VERBS" ;;
	esac
fi

if [ "${#subject}" -gt 72 ]; then
	violate "subject is ${#subject} characters; keep it at most 72"
fi
case "$subject" in
*.) violate "subject must not end with a period" ;;
esac

if [ -n "$second" ]; then
	violate "separate the subject from the body with a blank line"
fi

if [ -n "$violations" ]; then
	echo "commit rule violations:$violations" >&2
	exit 1
fi
