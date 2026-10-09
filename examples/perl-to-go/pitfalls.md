# Pitfalls: Perl → Go test ports

Generic entries from a real migration of ~260 Perl feature test files. Project-specific entries were removed. Format: **symptom** → cause → fix.

## Harness defaults

- **Times are off by 9 hours (or by your offset) only in the Go port.** The Perl factory sent a time zone header by default, so legacy test users had a time zone. The Go harness leaves it empty, and Go's `time.LoadLocation("")` returns **UTC without an error**, so the "fallback on error" path never runs. Set the time zone explicitly, with a reason comment.
- **A branch that "obviously" runs in production is skipped in the port.** The legacy harness sent a default locale, user agent, client version or network type. Diff the implicit request headers of both harnesses once, and record the result in the conversion map.

## Types

- **The Go assertion fails with `"1883" != 1883`, or the reverse.** Perl JSON types depend on where the value came from: DB accessor → string, literal or config → number. `cmp_deeply` compares loosely, so the legacy expected value proves nothing. Measure on the first run.
- **The same field is written as `'5+'` in one place and `1` in another in the `.t`.** Don't assume the Go types differ. Find the presenter/serializer that produces both fields. If one function produces them, the type is consistent.
- **A field expected to be 0 always passes.** It may be the string `"0"`, missing, or null. `.Uint()` returns 0 for all three. Assert the type (`Type == gjson.String`) together with the value.

## Dead legacy checks (strengthen and comment)

- `is $a->{user_id}, $b->{user_id}`, where neither response has a top-level `user_id`: undef equals undef, so it always passes. Map it to the field that really exists.
- `ignore` without a trailing comma swallows every following key/value pair as its arguments. Perl gives no warning. Many fields were never checked.
- `subbagof` with a list of expected elements passes on an empty result. One port found that the seed data had changed and the real result had been empty for a long time.
- A factory that calls APIs and ignores their status: the setup failed, and the test was still green. A fatal guard in Go makes the failure visible for the first time.

## Async and processes

- **Don't call a helper that can `t.Fatal` from inside a `require.Eventually` condition.** testify runs the condition in another goroutine. `FailNow` there violates the `testing` contract: the goroutine exits, and the real error is hidden behind "Condition never satisfied". Also, the timer takes priority over a slow condition that is still running, so a condition that takes seconds can time out after it has already succeeded. Write a deadline loop in the test goroutine instead.
- **Each call into the legacy stack is a new process.** Mocks, sessions and the mocked clock do not carry over between calls. If the legacy test wrapped "get → post → get → retry" in one mock scope, embed the whole sequence in one legacy call.
- **The legacy bridge reports success for a script that used Test::More.** `ok`/`is`/`cmp_deeply` inside a generated script without a plan exits with 254 ("no plan was declared"), or behaves unexpectedly. Use `die ... unless ...` in snippets. For large comparisons, `print encode_json(...)` and assert in Go. Cast types explicitly on the Perl side (`int($x)` / `''.$x`).
- **A web test client with auto-check disabled does not die on a 500 or a login redirect.** An error page parsed as JSON yields zero values, and a 0-count assertion passes. Guard with `require.True(t, gjson.Valid(res))`.

## Time

- **Mocked clock and real-time TTL.** A key computed from mocked time can expire on the wall clock, or not expire when the test expects it to. When the legacy test waited in real time, keep that wait in real time.
- **Operations through the bridge take seconds.** Anything that relied on "within the same second" in one Perl process becomes flaky. Widen windows or restructure; don't add sleeps.

## Lint (golangci-lint)

- **`unparam`:** an argument that every call site passes with the same value, or a return value every caller discards, gets flagged. This applies to test helpers too. It needs two or more call sites, so don't fold arguments of single-call helpers.
- **`gocritic hugeParam`:** triggers at 80 bytes or more. `gjson.Result` is exactly 80 bytes, and so is a struct with five string fields (16 × 5). Use pointers.
- **`gocritic commentedOutCode` misfires on non-English comments** that look like `identifier = value` or `identifier(...)`. Rephrase them as sentences.
- **`intrange` / `modernize`:** use `for i := range N`, not three-clause loops.
- **`testifylint empty`:** the `--fix` rewrites `assert.Equal(t, "", x)` to `assert.Empty`, and `Len(x, 0)` must be `Empty`. If you need to tell "missing" from "empty", assert `.Exists()` separately.
- **`gofmt`:** inside doc comments it turns `''` into `”`. Avoid quoting Perl's `''` in doc comments.
- **`misspell --fix`:** silently "corrects" typos you kept on purpose, such as test names that mirror legacy subtests. Suppress it on that line.
- **`go build` does not compile `_test.go` files.** Run `go vet` or the tests to type-check them. And "vet passed" is not "lint passed".
