# Review checklist (base)

Project-specific points live in the project profile's `review-checklist.md`; read both. Classify findings as `[must]` (must fix), `[imo]` (recommended), `[nits]` (trivial) or `[ask]` (question). Background for each point: `principles.md`.

## Coverage

- [must] **Every legacy test case is accounted for.** It is either ported, merged (name the target), or omitted with a reason recorded *in the new test file*. A missing case without a recorded reason is a finding.
- For "omit: covered", the substitute test exists and actually checks the same thing. Open it.
- Merged cases did not lose any of their checks.

## Assertion strength

- [must] **For each assertion, can you say what broken behavior it would catch?** Watch for the four shapes in `principles.md` §4: filter = expectation, a count that cannot change, half of an AND, and a baseline taken through a different path or actor.
- Every API call whose response is otherwise unchecked asserts at least its success indicator, **fatally**. Re-derive this for every call yourself; do not accept the plan's classification. The only exemptions are calls expected to fail and setup unrelated to the behavior under test. A call that drives the state transition under test is not "setup".
- Preconditions and values that later checks depend on use the fatal form.
- No assertion depends on an order the system does not guarantee.
- Negative checks (0 results, 404, "not visible") have a positive baseline through **the same actor and the same path**.
- Where the new language reads "missing" as a zero value, existence is asserted too.
- Expected values can be derived from the setup.
- A strengthening of a dead legacy check has a reason comment.

## Explanations

- Calls the legacy test did not make (extra daemon runs, fixed clocks, manual inserts) have a "why" comment.
- Explicit version, locale or header values that the legacy harness sent implicitly have a "why" comment.
- [must] **Every symbol, file path and constant named in a comment exists, and every stated cause holds in the implementation.** Trace it. Wrong causal comments pass CI and were the largest source of `[must]` findings.

## Code conventions

- Layout follows the project's rules (by default, one package per port). Fixtures are referenced, not duplicated.
- When appending to an existing package, new helpers have names specific to the feature, to avoid collisions with parallel PRs. **If you cannot tell which helpers are new, ask (`[ask]`); do not guess.**
- Comments explain only non-obvious "why". They do not reference the plan, knowledge files, review numbers, line numbers or the deleted legacy file.
- Project coding rules (see the project profile).

## Finishing (not for the reviewer: checked by Phase 5 before commit)

- The legacy test file is deleted.
- No profile or knowledge files are in the port diff.
