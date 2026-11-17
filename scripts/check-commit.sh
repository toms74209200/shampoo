#!/bin/sh
# Checks a commit against the commit rules.
#
# Usage:
#   check-commit.sh --msg-file <file>  check a message against the staged changes (commit-msg hook)
#   check-commit.sh --rev <rev>        check an existing commit (CI)
#
# Exit codes: 0 pass, 1 rule violation (task failure), 2 checker error (workflow failure).
#
# Rules:
#   1. Subject is ":<gitmoji>: <Verb> <what>", using the emoji and verbs listed below
#   2. Subject is at most 72 characters and does not end with a period
#   3. A body, if any, is separated from the subject by a blank line
#   4. A commit changes files of exactly one category (see category()).
#      A commit that must span categories states why in a trailer:
#        Cross-Category: <reason>

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

category() {
	case "$1" in
	.github/*) echo ci ;;
	.devcontainer/*) echo devcontainer ;;
	.githooks/* | scripts/* | Makefile | .golangci.yml | .gitignore) echo tooling ;;
	*.go | go.mod | go.sum) echo go ;;
	*.md | docs/* | LICENSE) echo docs ;;
	*) echo "other:$1" ;;
	esac
}

case "${1:-}" in
--msg-file)
	[ $# -eq 2 ] || workflow_failure "usage: check-commit.sh --msg-file <file>"
	[ -r "$2" ] || workflow_failure "cannot read message file: $2"
	msg=$(git stripspace --strip-comments <"$2") || workflow_failure "git stripspace failed"
	files=$(git diff --cached --name-only) || workflow_failure "git diff --cached failed"
	;;
--rev)
	[ $# -eq 2 ] || workflow_failure "usage: check-commit.sh --rev <rev>"
	msg=$(git log -1 --format=%B "$2") || workflow_failure "cannot read commit: $2"
	files=$(git diff-tree --no-commit-id --name-only -r --root "$2") || workflow_failure "cannot list files of: $2"
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

if [ -z "$files" ]; then
	violate "commit changes no files"
fi

categories=$(printf '%s\n' "$files" | while IFS= read -r f; do
	[ -n "$f" ] && category "$f"
done | sort -u)
count=$(printf '%s\n' "$categories" | grep -c .)
if [ "$count" -gt 1 ] && ! printf '%s\n' "$msg" | grep -q '^Cross-Category: ..*'; then
	detail=$(printf '%s\n' "$files" | while IFS= read -r f; do
		[ -n "$f" ] && printf '    %s: %s\n' "$(category "$f")" "$f"
	done)
	violate "commit mixes $count categories; split it into one commit per category, or add a trailer \"Cross-Category: <reason>\" if it cannot be split:
$detail"
fi

if [ -n "$violations" ]; then
	echo "commit rule violations:$violations" >&2
	exit 1
fi
