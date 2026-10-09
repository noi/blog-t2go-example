# test-porter

Claude Code skills for porting a legacy test suite to a new language during a language migration. **One legacy test file → one reviewed draft PR**, and every port makes the next one more accurate.

It grew out of a Perl → Go backend migration. There, 263 Perl feature test files were ported to Go E2E tests in about 9.5 weeks, in 251 PRs. In the two years before, the pace had been under two files a month. The workflow is language-agnostic. The knowledge is per project, and it ships with a Perl → Go starter pack.

[日本語版 README](README.ja.md)

## What it does

```
/port-test  <legacy test>            one file, with a plan approval gate
/port-test-batch  <a> <b> ... -j 2   many files, one worktree + PR each, J in flight
```

| Phase | Work | Who |
|---|---|---|
| 1 | Inventory every legacy case, classify it as port / merge / omit (with evidence), plan | session model |
| 2–3 | Implement, run, fix until green and lint-clean | `port-test-implementer` (sonnet) |
| 4 | Review: coverage, assertion strength, true comments | `port-test-reviewer` (opus, read-only) |
| 5 | Delete the legacy file, inspect the diff, commit, draft PR, feed knowledge back | session model |

Design points (details in [`principles.md`](skills/port-test/references/principles.md)):

- **Workflow and knowledge are separate.** The skill is generic. Your project's knowledge lives in `.claude/test-porter/` in *your* repo, and each port adds to it via a separate knowledge branch and PR, so parallel port PRs never conflict.
- **Each judgment goes to the model suited to it.** Planning and judgment use the session model, the volume of implementation work uses a cheaper model, and review uses a stronger model that has **no shell**. The coordinator feeds the reviewer the diff, test results and measured types.
- **Tests must not get weaker.** The implementer cannot add omissions, weaken assertions or change production code. It reports instead. Every omission needs evidence that another test covers the case.
- **Comments must be true.** In practice, the largest source of blocking review findings was wrong causal comments that pass CI. The reviewer is told to trace every stated cause.

## Install

As a plugin:

```
/plugin marketplace add <owner>/test-porter
/plugin install test-porter@test-porter
```

Commands are then `/test-porter:port-test` and `/test-porter:port-test-batch`.

Or copy `skills/*` into `.claude/skills/` and `agents/*` into `.claude/agents/` of your repository.

## Set up your project

```
/test-porter:port-test init
```

This scaffolds `.claude/test-porter/` (profile, conversion map, pitfalls, test execution, review checklist, PR template). If you are migrating Perl → Go, seed the knowledge files from [`examples/perl-to-go/`](examples/perl-to-go/).

Recommended start:

1. Fill in `profile.md`, especially **how to tell whether a component is still legacy or already migrated**. This decides which helper a port calls.
2. Port 2–4 files with close supervision. Mine their human review comments into `conversion-map.md` and `review-checklist.md`.
3. Use the single-file skill until the knowledge base stops growing quickly, then switch to batch mode with `-j 1`, then `-j 2`.

## Requirements

- Claude Code with subagents. Without them, the skill runs every phase in the main session.
- `git` with worktrees (batch mode) and `gh` (PR creation). If push is not allowed, the skill stops at the commit and prints the commands.
- A new test suite that can run against the system the legacy tests exercised, with isolation between parallel runs if you use `-j` > 1.

## License

MIT
