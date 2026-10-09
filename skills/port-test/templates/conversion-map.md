# Conversion map

Legacy → new mapping for this project. Seed it from `examples/<pair>/` if available, then grow it from real ports and review comments.

## Idioms

| Legacy | New | Notes |
|---|---|---|
| | | |

## Helpers

| Legacy factory / helper | New helper | Side effects to reproduce |
|---|---|---|
| | | |

## Mocks → real effects

| Legacy mock | Real sink in the test environment | How to read it |
|---|---|---|
| | | |

## Assertion conventions

- Success guard on otherwise unchecked calls: `<how, fatally>`
- Fatal vs non-fatal: `<project rule>`
- JSON / response comparison helper: `<name>` (and its gaps, e.g. "nil expectation cannot tell missing from null")

## File layout

- New ports: `<new_test_root>/<domain>/<name>/` as a new package (closes the namespace; parallel PRs cannot collide)
- Append to an existing file only when: `<rule>`
- Shared fixtures: reference, do not copy

## Reference examples

Ports worth imitating, closest-shape first:

- `<path>`: <what it shows>
