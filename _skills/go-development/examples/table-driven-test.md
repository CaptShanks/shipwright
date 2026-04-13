# Table-Driven Tests in Go

This document shows an idiomatic table-driven test layout: **subtests** with `t.Run`, **parallelism** where safe, **cleanup** with `t.Cleanup`, **helpers** with `t.Helper()`, and naming that reads well in failure output and `go test -run`.

## Example: testing a pure function

Suppose we implement a small normalizer used in APIs:

```go
package user

import (
    "errors"
    "strings"
)

// ErrEmptyUsername is returned when the input is empty after trimming.
var ErrEmptyUsername = errors.New("empty username")

// NormalizeUsername trims space, lowercases ASCII, and rejects empty results.
func NormalizeUsername(s string) (string, error) {
    s = strings.TrimSpace(s)
    if s == "" {
        return "", ErrEmptyUsername
    }
    return strings.ToLower(s), nil
}
```

### Test file: naming and structure

- Test file: `user_test.go` in the same package for white-box tests, or `user` with `package user_test` for black-box imports (`user.NormalizeUsername`).
- Test function: `TestNormalizeUsername` — name is `Test` + exported symbol under test.
- Subtest names: short, stable strings that encode the scenario (`empty after trim`, `ascii mixed case`). They appear in `go test` output and `-run` regexes.

### Full example

```go
package user

import (
    "errors"
    "testing"
)

func TestNormalizeUsername(t *testing.T) {
    t.Parallel() // safe: this test does not mutate shared package state

    tests := []struct {
        name    string
        in      string
        want    string
        wantErr error
    }{
        {
            name:    "lowercase unchanged",
            in:      "alice",
            want:    "alice",
            wantErr: nil,
        },
        {
            name:    "ascii mixed case",
            in:      "BoB",
            want:    "bob",
            wantErr: nil,
        },
        {
            name:    "trims surrounding space",
            in:      "  carol  ",
            want:    "carol",
            wantErr: nil,
        },
        {
            name:    "empty string",
            in:      "",
            want:    "",
            wantErr: ErrEmptyUsername,
        },
        {
            name:    "empty after trim",
            in:      "   \t  ",
            want:    "",
            wantErr: ErrEmptyUsername,
        },
    }

    for _, tt := range tests {
        tt := tt // capture range variable for parallel subtests (Go < 1.22 loop var semantics)
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            got, err := NormalizeUsername(tt.in)
            if tt.wantErr != nil {
                assertErrorIs(t, err, tt.wantErr)
                assertEqual(t, "username", got, "")
                return
            }
            assertNoError(t, err)
            assertEqual(t, "username", got, tt.want)
        })
    }
}

// assertEqual compares got and want; name describes the value for messages.
func assertEqual(t *testing.T, field string, got, want string) {
    t.Helper()
    if got != want {
        t.Fatalf("%s: got %q, want %q", field, got, want)
    }
}

func assertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func assertErrorIs(t *testing.T, err error, target error) {
    t.Helper()
    if !errors.Is(err, target) {
        t.Fatalf("error: got %v, want %v", err, target)
    }
}
```

**Note on loop variable capture:** Go 1.22+ fixes the classic `for _, tt := range` closure bug so `tt := tt` is no longer required for parallel subtests. Keeping `tt := tt` remains valid on older toolchains and in mixed-version CI—drop it when your module’s `go` directive guarantees 1.22+.

## `t.Helper()` — why it matters

Functions that wrap `Error` / `Fatal` should call `t.Helper()` at the top. That attributes failures to the **call site line** in the test body instead of inside the helper, which drastically speeds up debugging.

Use helpers for:

- repeated assertions (`assertJSONEqual`, `assertErrorIs`)
- fixture construction (see below)

Avoid helpers that hide **which subtest** failed unless the message includes `tt.name` or distinct context.

## `t.Cleanup()` — teardown order

Use `t.Cleanup` for resources tied to a test or subtest: temp directories, DB transactions, HTTP servers, etc. Cleanup runs in **LIFO** order after the test function returns (or fails).

### Example: temporary directory per subtest

```go
func TestWithFilesystem(t *testing.T) {
    t.Parallel()

    cases := []struct {
        name string
        // ...
    }{
        {name: "creates config file"},
    }

    for _, tc := range cases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            dir := t.TempDir() // Go 1.15+: auto cleanup, no t.Cleanup needed

            srv := startTestServer(t, dir)
            t.Cleanup(srv.Close)

            // ... exercise code under test ...
        })
    }
}
```

Prefer `t.TempDir()` over manual `os.MkdirTemp` + `t.Cleanup` when possible. Use explicit `t.Cleanup` for closing clients, rolling back transactions, or restoring environment variables:

```go
t.Cleanup(func() { _ = os.Unsetenv("FEATURE_X") })
```

## Parallelism rules of thumb

1. **Top-level** `t.Parallel()` in `TestXxx` is appropriate when the test does not mutate **package-level** state or global singletons.
2. **Subtest** `t.Parallel()` is appropriate when each subtest uses **its own data** (e.g. `tt` fields, `t.TempDir()`, local server on `:0`).
3. Do **not** call `t.Parallel()` when tests share a mutable database row, a global counter, or a map under test without synchronization.
4. Mixing parallel subtests with shared `setupTestDB(t)` that returns a single pool can be fine if tests use **transactions** rolled back in `t.Cleanup`, or isolated schemas—document the contract.

## Naming conventions

| Element | Convention |
|--------|------------|
| Test function | `Test` + name of exported API: `TestNormalizeUsername` |
| Examples (optional) | `ExampleNormalizeUsername`, `ExampleNormalizeUsername_secondForm` |
| Subtest name | Lowercase phrase, no spaces if you rely on `-run` regex simplicity; spaces are fine and readable: `t.Run("empty after trim", ...)` |
| Benchmark | `BenchmarkNormalizeUsername` |

Subtest names should be **stable**: CI dashboards and flaky-test trackers correlate on them.

## Running a subset

```bash
go test ./user -run TestNormalizeUsername/empty -count=1
```

The `-run` argument is a regex applied to the concatenation `TestName/Subtest/SubSubtest`.

## Brief checklist

- Table rows include `name` and all inputs/outputs; avoid magic shared state between rows.
- `t.Run(tt.name, ...)` for each row.
- `t.Helper()` in assertion helpers.
- `t.Cleanup` / `t.TempDir()` for teardown.
- `t.Parallel()` only when concurrency safety is obvious—prefer correctness over speed.
