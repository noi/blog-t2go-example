# PR body

```markdown
## Background

Migration of tests from <legacy> to <new>.

## Changes

Ported `<legacy test path>` to `<new test path>`.

| Legacy case | Result | New test / evidence |
|---|---|---|
| <name> | Port | <new test name> |
| <name> | Merge | <target test name> |
| <name> | Omit: covered | <file + test name> |
| <name> | Omit: obsolete | <reason> |

## Verification

CI.

## Notes

Generated with the port-test skill. <Differences from the legacy test and why.>
```

- Write the table for human reviewers; it can come straight from the plan.
- Title: `port: <legacy test path>` (full path; basenames collide).
