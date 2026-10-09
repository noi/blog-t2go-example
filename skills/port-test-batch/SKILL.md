---
name: port-test-batch
description: Port many legacy test files in one go. Creates one worktree, branch and draft PR per file, runs the port-test workflow with up to J files in flight, and feeds knowledge back once at the end.
argument-hint: "<legacy test> <legacy test> ... [-j <parallelism>] [-k <knowledge-branch>]"
user-invocable: true
disable-model-invocation: true
---

# port-test-batch: port many files

Targets: $ARGUMENTS

**The canonical porting procedure is the `port-test` skill and its references.** This skill is a coordinator that multiplies that procedure across N files; it does not restate the porting logic. When in doubt, read `port-test/SKILL.md` (`${CLAUDE_SKILL_DIR}/../port-test/SKILL.md`). Load the project profile exactly as port-test Phase 0 describes.

- One file = one worktree = one branch = one draft PR.
- `J` (`-j`) is the number of files **in flight at once**. Default 1 (sequential). There is no limit on the number of files.
- The user is asked only (a) at intake (B0) and (b) when a plan says the port itself is pointless (B2), or when an implementer reports that the plan must change (B3). Everything else runs to completion.

## Arguments

- `-j <N>`: parallelism. **Anything other than an integer ≥ 1 is an error** (`-j 0` would never free a slot).
- `-k <name>`: knowledge branch.
- A flag written with a space consumes the next word as its value.
- Every other argument is a target. Anything that does not match `legacy_test_glob` stops the run with an error.

| J | Guidance |
|---|---|
| 1 | Default. Lowest load and token use. |
| 2 | Safe in the original deployment (10 CPU / 8 GB container VM). |
| ≥3 | Memory pressure and leftover test data start to matter. **Confirm with the user first.** |

Test data isolation is the project's responsibility; `test-execution.md` should state whether parallel runs can interfere. J only trades host resources and tokens for wall-clock time.

## Naming

Remove the legacy test root prefix and the extension, then replace `/` and `_` with `-`:

```
<legacy_root>/a_b/c_d.<ext>  →  slug = a-b-c-d   branch = port/a-b-c-d   worktree = <WS>/port-a-b-c-d
```

```bash
ROOT=$(git rev-parse --show-toplevel)
WS=$(dirname "$ROOT")   # where worktrees live; if ROOT is the main clone, use where `git worktree list` keeps the rest
```

(`profile.md` may override the branch prefix.)

## B0: intake & preflight

Run every git command here with `-C "$ROOT"`.

1. `git fetch origin <base>`. Every check below is relative to `origin/<base>`.
2. Every target exists on **`origin/<base>`** (`git cat-file -e origin/<base>:<path>`), not just in the current worktree. If any target is missing, **stop the whole batch**. Do not run only part of it.
3. Fix J. If J ≥ 3, confirm with the user.
4. Slug collisions:
   - local and remote-tracking branches (`port/*`, `*/port/*`)
   - **`git ls-remote --heads origin 'port/*'`** (branches pushed from other machines are not in your tracking refs)
   - `git worktree list` and existing directories
   - **duplicates within the batch** (`a/b` and `a_b` produce the same slug)

   Confirm any collision with the user.
5. The local `<base>` is fresh (`git rev-list --count <base>..origin/<base>` is 0). Lint-diff tools usually compare against the local base, and a stale base makes the auto-fixer rewrite unrelated files.
6. **Choose the knowledge version** (`port-test/references/knowledge-loop.md`, "Versions"). Compare the profile in the current worktree, on `origin/<base>`, and on the knowledge branch. **Pass the chosen absolute paths to every agent.** Do the same comparison for this batch skill's project notes.
7. Run the environment preflight listed in `test-execution.md`: dependency stack running, disk space, leftover test data, tasks that must not run concurrently. **Anything destructive, such as cleanup or prune, needs user confirmation.** If the limits are hit during the batch, agents report and stop; they never clean up themselves.
8. Fix the knowledge branch (port-test knowledge loop, step 1). Fold this question into the summary.
9. **Ask once whether push and PR creation are allowed.** If they are not, the goal becomes "commit and print the commands" for every file. This is the only approval the summary asks for and waits on.
10. Present the summary: J, then `| # | target | slug | branch | worktree |`. Create one task per file and keep it updated.

## B1: create all worktrees

```bash
git -C "$ROOT" worktree add -b "port/<slug>" "$WS/port-<slug>" origin/<base>
# then copy the untracked setup files listed in profile.md (e.g. .env); stop that file if a required one is missing
```

Worktrees stay until their PRs merge. Check host disk space when N is large.

## B2: plans

Start one `port-test-planner` per file. Planners are read-only, so up to about 3 can run in parallel regardless of J. Each prompt gives:

- **the worktree as the base directory** for every investigation (target, layout check, substitute tests, routing config). Plans must use the same version the implementer will write to.
- **the profile paths chosen in B0** (absolute)
- "Follow port-test Phase 1, steps 1–7, and return the planner output format."

Then:

1. Save each plan to `<scratchpad>/port-test-batch/<slug>-plan.md`, **never inside a worktree**.
2. **Cross-check all plans.** This is the only point where all parallel PRs are visible together. Look for:
   - (a) two plans appending to the same existing file
   - (b) identical new package paths
   - (c) new helper names that collide across appending plans
   - (d) **several plans changing the same shared file** (common test server, factory)

   Fix collisions in the plans before B3.
   - **Early dispatch:** you may send a plan to B3 before the others return if it creates a **new package** and no plan still pending is in the same domain. If *every* plan creates a new package, same-domain plans are safe too, as long as they don't write the same shared file. Planner times differ by a factor of several, so waiting for all of them wastes J.
   - **Families:** when several files have the same shape, run one first and put its measured facts into the later prompts.
3. Present a summary table of the plans (target, output, port/merge/omit counts). Do not ask for approval.
4. **Exception:** if a plan questions the port itself ("everything is already covered", "cannot port yet"), ask the user about that file. Always offer **three options: skip / port anyway / delete-only PR**. A delete-only PR removes the legacy test and the fixtures only it uses, with a coverage table in the body. For "skip", remove the worktree and the branch, or they will collide next time.

## B3: rolling window of J files

Keep at most J files in flight. A file holds its slot until its Phase 5 finishes, including while a `[must]` is being fixed.

- J = 1: run each agent with `run_in_background: false`, explicitly.
- J ≥ 2: run in the background and free slots when **the Agent completion notification** arrives.
- **If no notification arrives for a long time (~60 min), check without disturbing anything:** the test containers or processes for the slug, the log file's mtime, changes in the worktree, and `ListAgents`. If the agent re-delegated, only its child may be running. Use SendMessage as a liveness probe only as a last resort, because it can break the notification you are waiting for.

### Implement + test (one launch)

Start `port-test-implementer`. Unlike single mode, **Phases 2 and 3 run in one launch.** The prompt contains:

- **At the very top, under "⚠ Read first": "Do not call the Agent tool at all (including fork). Do the work yourself."** Buried in the body, this rule was ignored. Placed first, re-delegation dropped from 1 in 9 files to 0 in the next 8.
- "Your working directory is `<worktree>`. Use absolute paths. Never touch another worktree."
- The profile paths chosen in B0, the target test, the output location, the full plan and the plan's path.
- **The exact test command form from `test-execution.md`, filled in for this worktree** (unique container or process names, the worktree's config). Hand over the whole command. A list of flags gets reassembled in the wrong order.
- **Prohibited:** stopping or removing shared infrastructure, and any prune or cleanup. Only resources the implementer started itself may be removed.
- "Run `<lint>` on the host until it reports zero findings. Then check `git status` for unrelated changes made by the auto-fixer. Report infrastructure problems instead of working around them: lint lock contention, a stale cache pointing to deleted worktrees, a stale base."
- "Do not copy causes from the plan into comments. Verify each one yourself, or write none." From the second file on, **list the kinds of wrong causes found in earlier files**.
- **Repeated review findings from earlier files, with their full text.** The knowledge files already contain them, but they are still missed.
- "Report knowledge candidates. Do not edit profile files."
- "Write logs to `<scratchpad>/port-test-batch/<slug>-test-<n>.log`, never inside the worktree."

If the implementer reports a contradiction with the plan, a need to change production code, or a need to weaken an assertion or add an omission, **ask the user about that file.** The exception: if the plan contradicts the project's conventions, **resolve it yourself in favor of the conventions.** Planners do misread conventions, and the implementer's report should not be squashed with "follow the plan".

### Review

Start `port-test-reviewer` the same way, foreground or background depending on J. Give it:

- the paths and the plan
- the new test cases and helpers
- **the `git status --porcelain` and `git diff` output, which you run yourself**
- test results, measured types and the fix history
- the focus line on comment causes, plus the cause errors seen in earlier files

Handling:

- **Verify each `[must]` against the real file before sending it back.** Checking is always cheaper than a round trip. Sometimes the reviewer's conclusion is right but its explanation is wrong. Never paste the reviewer's wording into the fix request; send the facts you verified.
- Send back *what* must be achieved, not *how*. One reviewer-proposed fix was impossible. The implementer proved it by measurement and proposed three alternatives; treat that as a result, not a failure.
- If the review quotes code or line numbers that do not match the file, it is based on a stale read. **Wait a few minutes for a second notification before acting.** Several notifications can arrive from the same reviewer; check which one matches the current file.
- Decide `[imo]` / `[ask]` yourself and record each decision and its reason for the final report.

### Finish (coordinator)

**First confirm that the implementer's report shows every test passing.** Otherwise handle the file as a failure.

Then follow port-test Phase 5, steps 1–4, inside the worktree (`git -C <worktree>`; for `gh`, use `bash -c 'cd <worktree> && …'`). **Never skip the `git status --porcelain` inspection.** Remember that `git checkout --` does not remove untracked files, and `git add -A` will commit them.

If push is blocked, commit, record the commands, **keep the queue going**, and report everything at the end.

### Failures

- **A rate limit or session limit is not a failure.** If the worktree is clean (`git status --porcelain` is empty), resubmit the same prompt **after the reset time**; resubmitting earlier fails the same way. Several agents fail together when this happens.
- Any other failure: free the slot, keep the worktree and branch, and report the phase and cause. Note that the leftovers must be removed before the file is retried.

## B4: aggregate & knowledge

1. Result table: `| target | branch | PR | port/merge/omit | review decisions |`, with failures listed separately.
   - **Cross-check the helper names each implementer added** against each other. B2 only saw the plans, and two PRs that each build alone can still fail to compile together once both merge.
2. **Feed knowledge back once, for all files.** Deduplicate the candidates, route each to the right file (port-test knowledge loop), and **look for corrections to existing entries** as well as additions. Before committing, check the knowledge branch's current HEAD against its remote; other sessions may have pushed in the meantime.
3. Do not remove worktrees. Print the cleanup commands for after the merge.

## Parallel safety (summary)

- One worktree per concurrent test run. Never run two tests from the same worktree.
- Shared infrastructure is never stopped or pruned by agents.
- Scratch files never go inside worktrees.
- Unapproved push and PR edits apply only to the PRs this batch creates.
