# Conversion map: Perl (Test::More / Test::Deep) → Go (testing / testify / gjson)

Seed for a project's `conversion-map.md`. This covers only language-level and library-level mappings. Project helpers (test server, factories, legacy bridge) are added by each project.

## Test structure

| Perl | Go | Notes |
|---|---|---|
| `subtest 'name' => sub { ... }` | `t.Run("name", func(t *testing.T) { ... })` | Keep the original subtest name, even with typos (suppress `misspell` locally). |
| `done_testing` | (implicit) | |
| `local $ENV{...}` / `mock_guard` scope | setup + `t.Cleanup` | Mock scopes become observation windows; see "Mocks". |
| per-subtest DB reset | Often **not** per `t.Run` | If the new harness loads fixtures once per test, rows from all subtests coexist. Separate the data or the fixture directories. |

## Assertions

| Perl | Go | Notes |
|---|---|---|
| `ok $x` | `require.True` / `assert.True` | Prefer an explicit value comparison; `ok` is weak. |
| `is $got, $exp` | `assert.Equal` / `require.Equal` | `is` compares as strings. Decide the Go type by measurement (pitfalls: "Types"). |
| `is $x, undef` | `Type == gjson.Null` **and** `.Exists()` | gjson returns a zero `Result` for a missing key, and its `Type` is `Null`. |
| `cmp_deeply $got, $hash` (plain hashref) | exact match of all keys + fixed key count | Plain hashrefs compare the **key set** exactly. |
| `superhashof(...)` | path-based partial match | Only the listed keys, but **each must exist**. |
| `bag(...)` | unordered match keyed by id + `require.Len` | Unordered, but the count is exact. Indexed paths (`x.0`, `x.1`) would silently fix the order. |
| `subbagof(...)` / `subsetof(...)` | exact match against the **measured** set | Means actual ⊆ expected, so **an empty result passes**. Don't translate it to `Contains`. |
| `superbagof(...)` | `Subset` + a reason | The direction is the reverse of `subbagof`. |
| `ignore()` | `NotZero` / `Exists()` | Skips the value but **still requires the key**. Dropping the field weakens the test. |
| `like $x, qr/.../` | `assert.Regexp` or exact match | Check that the regex distinguishes the branches; a shared prefix proves nothing. |
| `dies_ok` / `throws_ok` | `require.Error` + error type/message | |

## Responses (JSON)

- Perl APIs often return **HTTP 200 with a body-level status** (`status.ok = 0`) on error. Check the body status too, not just the HTTP code.
- If the call's response is otherwise unchecked, assert the success flag **fatally**: `require.EqualValues(t, 1, gjson.Get(res, "status.ok").Int())`.
- gjson:
  - A missing key returns a zero `Result`. `.Int()`, `.String()` and `.Bool()` then return zero values. Pair them with `.Exists()` when "missing" would be a bug.
  - `path.#` returns 0 when `path` is missing. A 0-count assertion therefore needs a success guard.
  - `path.#(cond)` returns only the **first** match. Use `#(cond)#` to count matches.
  - `gjson.Result` is 80 bytes, and gocritic `hugeParam` triggers at 80 or more. Pass it to helpers by pointer.
- Before using an ID extracted from a response, guard it with `require.NotEmpty`. A parse failure returns 0, and many factories treat 0 as "create a new one".

## Mocks → real effects

| Perl pattern | Go |
|---|---|
| Mocked enqueue, accumulated array, `cmp_deeply` against the whole array | drain → act → drain on the **real** queue. Fix the window's total count with `require.Len`, **and** assert 0 of the relevant kinds during setup. |
| Mocked push / logger | Read the real sink, if the test environment has one. |
| `dequeue(sub { is undef, shift })` | Watch out: some queue helpers **do not call the callback on an empty queue**. Then the next `cmp_deeply $row` compares the previous row again. Replace it with drain-and-count. |

## Time

| Perl | Go | Notes |
|---|---|---|
| `set_fixed_time($t)` | the project's mocked clock | Always an absolute offset. |
| `my $time = time;` inside a helper | — | It **shadows** the outer `$time` and does not advance it. Don't make the Go helper return "the new time" for the next phase to start from. |
| `Sub::Retry::retry($times, $delay, \&code, \&retry_if)` | deadline loop in the test goroutine | 3rd argument = code, 4th = condition. Don't use `require.Eventually` if the condition can call `t.Fatal` (see pitfalls). |

## Encoding

- Perl test files usually have `use utf8;`. Generated Perl snippets that a Go bridge executes **often don't**. Without it, non-ASCII literals are byte strings:
  - Values written to the DB are double-encoded.
  - Regex matches against decoded text never match. **Negative matches therefore always pass.**
- `print` of non-ASCII without `binmode STDOUT, ':utf8'` prints "Wide character" warnings and no reliable UTF-8 output.
- JSON responses may contain `\uXXXX` escapes, so matching non-ASCII text against the raw body fails. Decode the JSON first.

## Strings embedded into generated Perl

Use **single quotes** for values. Inside double quotes, `@x` and `$x` are interpolated, and an undefined variable silently becomes an empty string. Use double quotes or `join` only where you need `\n` / `\r\n`.

## Determinism

- Perl hash order is randomized per process. Repository code that does `SELECT ... LIMIT 1` without `ORDER BY` against rows a factory inserted by iterating a hash is non-deterministic. Make the rows unique or identical.
- Go `map` iteration is randomized too. Never assume an order.
