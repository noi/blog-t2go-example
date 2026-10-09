# test-porter profile

Project-specific settings for the `port-test` / `port-test-batch` skills. Edited by humans; the knowledge files next to it grow automatically.

## Migration

- Language pair: <legacy> → <new>            (e.g. Perl → Go)
- legacy_test_glob: `<t/**/*.t>`
- new_test_root: `<tests/>`
- Project testing rules to read: `<path or "none">`

## Legacy or migrated?

How to tell whether a component the test touches is still implemented in the legacy stack:

- Endpoints: <e.g. "look up the path in the routing config `<file>`; listed under the new cluster = migrated, listed under legacy *or not listed* (catch-all) = legacy">
- Workers/daemons: <how to tell>
- Helpers to call for each: <e.g. `s.Post` for migrated, `s.PostLegacy` for legacy; `s.RunLegacy` only for running legacy-only workers>

## Commands

- Run one new test package: see `test-execution.md`
- lint: `<make lint-diff>`   (must report 0 findings; this is the CI entry point)
- base branch: `<main>`

## Git / PR conventions

- Single-run branch: `<user>/port-<short-name>`
- Batch branch prefix: `port/`
- Commit message / PR title: `port: <legacy test path>`
- PR label: `<port>`        Knowledge PR label: `<port-knowledge>`
- Knowledge branch name pattern: `<*port-knowledge*>`
- Untracked files to copy into new worktrees: `<.env, .claude/settings.local.json>`

## Batch notes

<Coordinator-level knowledge discovered in batches: safe J, resource limits, notification quirks, …>
