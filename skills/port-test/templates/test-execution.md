# Test execution

## Run one package

```bash
<command that runs one new test package against the shared dependency stack>
```

## Long runs

<How to start a long run detached and poll it in the foreground, e.g. nohup + a bounded wait loop. Background-shell completion notifications are unreliable after an agent is resumed.>

## Parallel runs (batch)

- Isolation between concurrent runs: <e.g. per-test DB namespace>
- Per-worktree command form: <exact command with unique container/process name>
- Must not run concurrently with: <e.g. unit tests that wipe a shared index>

## Preflight (batch B0)

- Dependency stack running: `<command>`
- Disk space: `<command>`, ask before cleanup below `<threshold>`
- Leftover test data: `<command>`, ask before cleanup above `<threshold>`

## Troubleshooting

<lint lock contention, stale lint caches pointing at deleted worktrees, …>
