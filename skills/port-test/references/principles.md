# Porting principles

Language-agnostic lessons for porting tests from a legacy stack to a new one. Every item here came from a real port that went wrong first. Project-specific knowledge belongs in the project profile (`conversion-map.md` / `pitfalls.md`), not here.

Read this in Phase 1 (planning) and before implementing. Reviewers read `review-checklist.md`, which points back here.

## 1. A port is a re-verification, not a transliteration

- **The goal is "the new test fails whenever the old one should have", not "the new code looks like the old code".** Port the *intent* of each check. Transliterating line by line copies the old test's blind spots and adds new ones from semantic gaps between the two test frameworks.
- **Assume the old test may already be dead.** Common forms:
  - A dequeue/poll helper that silently does nothing on an empty queue, so the assertion compares the *previous* value again.
  - A matcher that passes on empty input (e.g. "is a subset of" against `[]`).
  - A syntax slip that swallows the following checks (a missing comma after a wildcard matcher in a nested structure).
  - A regex that only matches a prefix shared by both branches it is supposed to distinguish.
  - A loop whose body never runs.
  - A validator that returns the same error message for several branches, so matching the message does not prove which branch ran.

  When you find one, strengthen the check in the port **and leave a one-line reason in the code**. Without it, the next person deletes the "extra" assertion.
- **Old fixtures may not have checked their own success.** A factory that ignores the return value of the API calls it makes lets the old test stay green while the setup actually failed. If "the old test passes but the new one fails", first check whether the old setup really succeeded.

## 2. Classify every legacy case, with evidence

Each legacy test case (subtest / `it` block / test function) gets exactly one class:

| Class | Meaning | Required evidence |
|---|---|---|
| Port | Re-implemented in the new test | New test name |
| Merge | Folded into another new test case | Target test name |
| Omit: covered | Already verified elsewhere | **File and test name of the existing test, confirmed to exist and to check the same thing** |
| Omit: obsolete | The test checks nothing meaningful anymore (route removed, feature gone, dead assertion) | Concrete reason |

- An omission recorded only in the PR description is lost once the legacy file is deleted. **Record omissions in the new test file as well**, so a future reader can tell a deliberate omission from a missed port.
- If every case is "omit: covered", the best outcome may be **a PR that deletes the legacy test without writing a new one**. Offer this as a third option next to "port anyway" and "skip".

## 3. Prefer observing real effects over mocks

- Legacy tests often mock an outbound boundary (a queue, push notification, logger, external API) and inspect the mock's recorded arguments. In the port, **read from the real sink when one exists in the test environment** (dequeue from the real queue, read the real table) and verify the payload.
- Reproduce mock *scope* as a window: drain → act → drain, and assert only on what appeared inside the window.
- **Do not throw away what happened during setup.** A legacy assertion of exact equality on an accumulated mock log also proved that *setup produced nothing unexpected*. If your window discards the setup period, assert explicitly that the relevant message kinds did not appear during setup.
- During a migration, a mock written in the legacy language **cannot see effects produced by components already moved to the new language**. Some legacy tests were "fixed" by changing the expected count to 0. Reading the real sink lets you restore the real expectation.
- If you replace a mock by seeding real data, check that the code really reads the place you seeded. A write to the wrong table is silent.

## 4. Assertion design: "what would make this fail?"

For every assertion, you should be able to say what broken behavior it would catch. If you cannot, the assertion is decoration. Four recurring shapes:

1. **Filter equals expectation.** `SELECT ... WHERE id IN (3,4,5)` and then expect `{3,4,5}`. Over-granting outside the filter is invisible.
2. **Count that cannot change.** Asserting "did not increase" on something the design aggregates into one row.
3. **Half of an AND.** The behavior needs A and B, and the test only shows "0 results when A is false". Add the case where only the other side is false.
4. **Baseline from a different path.** A negative check ("not visible", 404, empty list) needs a **positive baseline through the same actor and the same path**. A 404 also happens on a typo or a failed precondition. Show that the same request returned 200 before the deletion.

Further rules:

- **Fatal vs. non-fatal:** use the fatal form (stop the test) for preconditions and for values that later checks depend on. Use the non-fatal form only where it is useful to see several independent failures at once.
- **Minimum success guard:** when an API call's response is otherwise unchecked, assert at least its success indicator, and assert it fatally. This applies even when the legacy test checked nothing.
- **Order:** do not assert an order the system does not guarantee (queues, unordered sets). Count by kind instead.
- **Zero values:** in the new language, "missing" often reads as a zero value (`""`, `0`, `false`, `nil`). Assert existence as well as value when "missing" would be a bug.
- **Derivable expectations:** expected values should be derivable from the setup, not magic numbers.

## 5. Types and defaults are facts to measure, not to infer

- **Implicit defaults differ between test harnesses.** The legacy harness may have sent a locale, time zone, user agent, network type or client version implicitly. The new harness sends something else, or nothing. "Not specified" is not "does not depend on". Where you set a value explicitly, write one line explaining why.
- **Empty is not absent.** An empty string can take a different branch than a missing key. Some standard libraries silently map empty input to a default (for example, an empty time zone name meaning UTC).
- **Dynamic-language response types depend on where the value came from** (database → string, literal → number) and may differ across endpoints for the same field. The legacy expected value, which was compared loosely, is not evidence of the type. Neither is the API spec. **Determine the type by measurement on the first run**, then assert type and value together.

## 6. Time, processes and hidden daemons

- **Mocked clock and real-time TTL do not mix.** Keys may be computed from mocked time while expiry follows wall-clock time.
- **Crossing a process boundary changes timing.** Operations that took milliseconds inside one legacy process can take seconds when the new test calls into the legacy stack through a bridge. This shows up as flakiness only in CI.
- **In-process caches cannot be flushed from outside.** Reorder the test instead of fighting the cache.
- **Do not add daemons or workers the legacy test did not run.** Processing a queue that the legacy test left alone changes the preconditions, and assertions that used to pass start failing. Leave a comment saying why you did *not* add it, or someone will.
- The reverse also happens: in the legacy environment some async work ran synchronously in "test mode". The port may need to run that worker explicitly.

## 7. Comments: few, true, and self-contained

- **Write only non-obvious "why" comments.** Do not restate what the function name and code already say. A 350-line test with 60 lines of comments was judged excessive in review.
- **Never reference things outside the repository**: the conversion plan, knowledge files, review comment numbers. Write the reason itself in one line.
- **Never reference line numbers, and never reference the legacy test path**, which is deleted when the port merges.
- **A wrong causal comment passes CI.** This was the largest single source of blocking review findings: in one batch, every `[must]` was a false explanation in a comment. Typical errors:
  - Naming a function that does not exist as the source of a behavior.
  - Reversing the direction of a defaulting operator.
  - Claiming a branch "can fire" when it cannot be reached in real time.
  - Attaching a reason to two lines when it applies to only one.
  - Copying a plausible but wrong cause from the plan.
- **Rule: only write a cause you have verified in the implementation yourself.** A value with no explanation is better than a wrong explanation. Limit causal comments to values that cannot be read from the code.

## 8. Parallel work and namespaces

- Grep cannot see unmerged parallel PRs. **Give each port its own package or module** so that helper names cannot collide between PRs that are open at the same time.
- Shared helper files, such as the common test server and factory, are the one place collisions remain. Add to them, never rewrite them, and coordinate ordering when several ports touch the same file.
- **Run one port of a family first**, then give its measured facts (types, response shapes, which risks turned out not to matter) to the later ports. In practice the later ports then pass on the first try.

## 9. Tooling

- **"Passed vet/fmt" is not "passed lint".** Run the same lint entry point CI uses, on the host where it works, and get it to zero findings before review.
- Auto-fixing linters may rewrite unrelated files when the diff base is stale. Inspect `git status` after every auto-fix run.
- A linter can "correct" intentional content: typos kept on purpose, quotes inside doc comments, code-like prose in comments. Suppress these locally with a reason; do not let the fixer change the meaning of the test.
