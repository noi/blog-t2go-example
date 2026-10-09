---
name: port-test
description: Port one legacy test file to the new language's test suite during a language migration (e.g. Perl → Go). Runs plan → implement → test → review → finish with model routing, and ends with a labeled draft PR plus a knowledge-base update.
argument-hint: "<legacy test file> [-k <knowledge-branch>] | init"
user-invocable: true
disable-model-invocation: true
---

# port-test: port one legacy test to the new stack

Target: $ARGUMENTS

One run handles **one legacy test file = one PR**. The goal is a **draft PR with the project's port label**, plus any new knowledge fed back into the project profile.

This skill is the *workflow*. What the project knows about its own migration lives in the **project profile**, a directory in the target repository (default `.claude/test-porter/`):

| File | Content | Grows by |
|---|---|---|
| `profile.md` | Paths, commands, branch/label conventions, how to tell legacy from migrated components | Edited by humans |
| `conversion-map.md` | Legacy → new idiom map, helper map, assertion conventions, file layout, reference examples | Knowledge loop |
| `pitfalls.md` | Pitfalls actually hit (only real ones) | Knowledge loop |
| `test-execution.md` | How to run the new tests locally, in parallel, and how to debug | Knowledge loop |
| `review-checklist.md` | Project-specific review points (added to the bundled checklist) | Knowledge loop |
| `pr-template.md` | PR body format | Edited by humans |

Bundled, project-independent references (read-only):

- `${CLAUDE_SKILL_DIR}/references/principles.md`: porting principles (read in Phase 1 and by every agent)
- `${CLAUDE_SKILL_DIR}/references/review-checklist.md`: base review checklist
- `${CLAUDE_SKILL_DIR}/references/knowledge-loop.md`: how to feed knowledge back (Phase 5, step 5)
- `${CLAUDE_SKILL_DIR}/templates/`: profile templates for `init`

## Arguments

- `init`: scaffold the project profile (see "Init" below) and stop.
- `-k <name>` / `-k<name>`: knowledge branch name. If omitted, ask in Phase 5.
- A flag written with a space **consumes the next word as its value** (`-k foo`: `foo` is not a file).
- Every other argument is the target legacy test path. If it does not match `legacy_test_glob` in `profile.md`, **stop with an error**. Do not silently reinterpret a typo.
- No target: ask the user.

## Phase 0: load the profile

1. Find the profile: `$CLAUDE_PROJECT_DIR/.claude/test-porter/profile.md`, or the directory named in the project's `CLAUDE.md` under "test-porter profile".
2. If it is missing, offer to run **Init**. Do not guess project conventions.
3. Read `profile.md` fully. All `<placeholders>` below (`<base>`, `<label>`, `<lint>`, …) come from it.

## Phases and model routing

| Phase | Work | Who |
|---|---|---|
| 1 | Inventory & conversion plan | Main session (session model) |
| 2 | Implement | Agent `port-test-implementer` (sonnet) |
| 3 | Run tests & fix loop | Same implementer, continued via SendMessage |
| 4 | Review | Agent `port-test-reviewer` (opus, read-only) |
| 5 | Finish, PR, knowledge loop | Main session |

- Models are set only in the agents' frontmatter (aliases such as `sonnet`/`opus`, never full model IDs). To change the routing, edit those files only.
- Phases 1 and 5 hold the judgment, so they follow the session model (`/model`).
- When installed as a plugin, agents are namespaced (`test-porter:port-test-implementer`). If the agents are unavailable, run every phase in the main session, reading the references directly. **The quality of judgment comes from the references, not from the agents.**
- Agents do not inherit this skill's directory. **Always pass absolute paths** to the profile and to `${CLAUDE_SKILL_DIR}/references/*` in agent prompts.

## Phase 1: inventory & plan (main session)

1. Read the target legacy test closely and list **every** test case.
2. For each component the test depends on (endpoint, daemon/worker, factory), decide whether it is **legacy or migrated**, using the method in `profile.md` (routing config, service registry, …). This determines which helper to call (e.g. "call the new handler" or "call through the legacy bridge"). Use the legacy bridge only for components still implemented in the legacy stack. Do not use it as a shortcut for setup.
3. Check legacy factory side effects. Decide whether each one represents real production behavior that the new setup must reproduce.
4. Classify each case: **port / merge / omit: covered / omit: obsolete** (`principles.md` §2). For "omit: covered", **open the substitute test and confirm that it exists and checks the same thing.**
5. Decide how mocks become observations of real effects (`principles.md` §3, `conversion-map.md`).
6. Read `pitfalls.md` and work the relevant items into the plan.
7. Decide the output location using the layout rules in `conversion-map.md`. Default: a **new package per port**. Append to an existing file only when the rules say so.
8. Present the **conversion plan** and get approval before Phase 2. The plan contains: the classification table, legacy/new mapping, how differences are absorbed, the output location, and open questions. If the user's environment may forbid pushing, ask about push permission now, in one line.

## Phase 2: implement (port-test-implementer)

Start the implementer agent. The prompt must contain:

- The absolute paths of the target test, the output location, the profile directory, and `${CLAUDE_SKILL_DIR}/references/principles.md`.
- The full approved plan.
- "If you find a fact that contradicts the plan, do not absorb it in the implementation. Report and wait."
- "Do not copy causal explanations from the plan into code comments. Verify them in the implementation first, or leave them out."

## Phase 3: test & fix loop (same implementer)

Continue the same implementer via SendMessage, following `test-execution.md`.

- Any fix that changes the plan must come back to the main session: adding an omission, weakening an assertion, or changing production code. Ask the user if in doubt.
- Done means **all tests pass and the project's lint entry point (`<lint>`) reports zero findings.**

## Phase 4: review (port-test-reviewer)

Start the reviewer with: the target test path, the new file paths, the plan, and the names of newly added test cases and helpers (essential when appending to an existing file).

- **The reviewer has only Read/Grep/Glob.** Do not ask it to run `git diff`. Run it yourself and paste the output. Pass new untracked files as paths.
- **Feed it what it cannot obtain itself:** test results, types confirmed by measurement, the history of fixes. Without these it can only say "verify against the real response".
- Tell it that stray files in the diff are out of scope; Phase 5 checks those.
- Always include the focus line: "check that every symbol, path and constant mentioned in comments exists, and trace each stated cause into the implementation". This is the main defense against wrong-but-green comments (`principles.md` §7).
- `[must]`: **verify it against the actual file first**, then send it back to the implementer. State *what* must be achieved and let the implementer choose *how*. A fix the reviewer proposes can turn out to be impossible.
- `[imo]` / `[ask]`: decide in the main session; ask the user when unsure.
- If the review quotes code or line numbers that do not match the file, it was based on a stale read. Wait for a second notification before acting.

## Phase 5: finish & PR (main session)

Steps 1–4 run without asking for approval. **This exemption covers only the port PR this skill creates.** It does not extend to any other commit or push.

1. Delete the legacy test file (and fixtures used only by it; grep to confirm).
2. Run `<lint>` until it reports zero findings.
3. Commit on a branch named per `profile.md`. If you are on `<base>`, create the branch first.
   - **Before committing, inspect `git status --porcelain`.** This is the only checkpoint before an unapproved push. Check:
     - (a) Only the files the plan expects are present.
     - (b) **The legacy test deletion is included.**
     - (c) No profile files are included; knowledge goes on its own branch.
     - (d) The lint auto-fixer did not touch unrelated files.

     Revert anything unexpected. Tracked files: `git checkout --`. Untracked files: `rm`.
   - The commit message equals the PR title in `profile.md` (default `port: <legacy test path>`).
4. Write the PR body per `pr-template.md` into the scratchpad. Push, then run `gh pr create --draft --base <base> --label <label> --title ... --body-file ...`. Report the URL.
   - **If push or PR creation is blocked by permissions, do not retry or work around it.** Stop at the commit and report the branch name and the exact commands to run.
   - Do not add "ready for review" labels; the human does that when un-drafting.
5. **Knowledge loop:** if anything new was learned, follow `${CLAUDE_SKILL_DIR}/references/knowledge-loop.md`. Knowledge goes to a separate knowledge branch and never into the port PR. Skip this step if nothing new was learned.

## Init

Scaffold `.claude/test-porter/` from `${CLAUDE_SKILL_DIR}/templates/`:

1. Ask for the language pair and the legacy/new test locations. Inspect the repository to propose values (test runner, lint command, base branch, routing config) rather than asking blindly.
2. Copy the templates and fill in `profile.md`.
3. If the plugin ships an example for the language pair (`${CLAUDE_SKILL_DIR}/../../examples/<pair>/`), offer to seed `conversion-map.md` / `pitfalls.md` from it.
4. Recommend porting **2–4 files by hand or with close supervision first**. Then mine their review comments into `conversion-map.md` and `review-checklist.md`; that is the best initial knowledge base.

## Batch mode

`port-test-batch` runs this workflow over many files in parallel worktrees. Where it gives different instructions, **the batch skill wins**. The main differences: the plan comes from `port-test-planner` and has no approval gate, Phases 2 and 3 run in a single launch, branches are `port/<slug>`, and agents report knowledge candidates instead of editing the profile.
