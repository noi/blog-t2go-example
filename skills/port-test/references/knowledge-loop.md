# Knowledge loop

The skill stays accurate only because every port feeds what it learned back into the project profile. In the original deployment, the pitfalls file grew from 34 to 523 lines over about 250 ports. The workflow skill itself barely changed. Accuracy came from the knowledge base.

## What goes where

| Learned | File |
|---|---|
| A trap you actually stepped in (symptom → cause → fix) | `pitfalls.md` |
| A legacy → new mapping, assertion convention, layout rule | `conversion-map.md` |
| How to run, parallelize, clean up, lint | `test-execution.md` |
| A point reviewers keep raising | `review-checklist.md` (project) |
| How the batch coordinator should behave | the project's batch notes in `profile.md` |

Rules for entries:

- **Only things you actually hit.** Do not record speculation.
- Each entry gives the symptom you would search for, the cause, and the fix. Name the general shape, then the concrete example.
- **Look for corrections, not just additions.** Knowledge files can be wrong too. A wrong entry gets copied into code comments and becomes a review finding. When you correct an entry, say what was wrong ("an earlier version of this entry claimed X; that does not hold because Y").
- Revisit stale notes, such as "until PR X merges" or dated TODOs, whenever you touch a section.

## Why a separate branch

Knowledge edits **never go into the port PR**:

- Many port PRs are open at once. If each one edits `pitfalls.md`, they all conflict.
- Port PRs are reviewed for test correctness. Mixing in doc changes dilutes that review.

So knowledge accumulates on one **knowledge branch** with one long-lived draft PR. Each run adds commits to it.

## Procedure

1. **Pick the branch.**
   - If `-k` was given, use it.
   - Otherwise list the candidates with **both** `git branch -a --list '*<knowledge-pattern>*' --sort=-committerdate` **and** `git ls-remote --heads origin '*<knowledge-pattern>*'`. Without `ls-remote`, you miss branches pushed from other machines or sessions, and your push later fails as non-fast-forward. Exclude port branches (`port/*`, `*/port/*`). Ask the user which branch to use or what to name a new one.
2. If the branch is checked out in another worktree (`git worktree list`), edit and commit there; `git switch` would fail.
3. **Before switching, run `git status --porcelain -- <profile dir>`.** Uncommitted profile edits on the current branch would otherwise be carried over silently and committed onto the knowledge branch. Stash them away under a unique name first.
4. Commit and push:

   ```bash
   git switch <knowledge-branch>                 # or: git switch -c <name> origin/<base>
   # edit the profile files
   git add <profile dir>
   git commit -m "[skip ci] port-test: <summary of what was learned>"
   git push -u origin <knowledge-branch>
   git switch -                                  # back to the port branch
   ```

   - **Never force-push.** On a non-fast-forward rejection, stop and report; someone else may be stacking knowledge on the same branch.
   - If push is blocked by permissions, stop at the commit and print the commands.
5. If no open PR exists for the branch, create a draft one with a fixed title (`port-test: knowledge update`) and the knowledge label. If one exists, push only; do not rewrite its body.

## Versions: which knowledge does an agent read?

New worktrees are created from `origin/<base>`, so **they do not contain unmerged knowledge**. The coordinator's own branch may be hundreds of commits behind. Before launching agents, compare the profile in three places: the current worktree, `origin/<base>`, and the knowledge branch. Choose the newest version, and patch in any section that exists only in one of the others. **Pass that version's absolute path to every agent.** In one batch this step was skipped, and every agent ran with knowledge that was missing 77 lines already on `origin/<base>`.
