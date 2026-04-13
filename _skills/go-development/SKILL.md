---
name: go-development
description: >-
  Go development idioms, patterns, and conventions for writing idiomatic,
  production-quality Go code. Use when writing, reviewing, or refactoring Go
  code, or when setting up Go project structure, testing, and error handling.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---

# Go Development

This skill encodes how experienced Go teams write code that is readable, safe under concurrency, easy to test, and stable across releases. Go favors **simplicity, explicitness, and composition** over cleverness.

## Accept Interfaces, Return Structs

**Rule of thumb:** Functions should accept the **smallest interface** that expresses what they need, and return **concrete types** unless abstraction is required.

### Why accept interfaces

- Callers can pass mocks, adapters, or alternate implementations without widening your API.
- You document dependencies by method set, not by concrete package coupling.
- Prefer `io.Reader` over `*os.File`, `context.Context` everywhere for cancellation, and tiny local interfaces (`type Validator interface { Validate() error }`) at the point of use.

### Why return structs (concrete types)

- Callers get a clear, stable type with fields and methods they can rely on.
- Returning a large interface hides capabilities and encourages interface pollution upstream.
- Return an interface only when you have **multiple implementations** that are genuinely interchangeable at that boundary (e.g. `fs.FS`, `http.RoundTripper`), or when the type is intentionally opaque (`error`, sometimes `Reader` from a factory).

### Interface definition site

- Define interfaces **where they are consumed**, not where types are implemented. The standard library does this consistently.
- Avoid "preemptive" interfaces: if only one concrete type exists, a package-level `Fooer` interface often adds indirection without benefit.

## Error Wrapping and Propagation

Go treats errors as **values**. Professional code preserves causality and allows callers to branch with `errors.Is` / `errors.As`.

### Wrapping at boundaries

When crossing a logical layer (HTTP handler → service → repository), wrap with context the next layer lacks:

```go
if err != nil {
    return fmt.Errorf("load user %q: %w", userID, err)
}
```

Use `%w` only when the wrapped error should remain **inspectable** (`errors.Is` / `errors.As`). Use `%v` when the underlying error is an implementation detail you do not want callers to match on.

### Custom errors

Use sentinel errors (`var ErrNotFound = errors.New("not found")`) or small types implementing `Error()` when callers must distinguish cases. Document which errors are stable API.

### Logging vs returning

Do not log and return the same error at every layer; that duplicates noise. Typically: **return wrapped errors** from libraries and services; **log once** at the edge (handler, worker main) where you have request scope and tracing.

### `panic` and `recover`

Reserve `panic` for **programmer mistakes** and invariant violations that should crash tests and development builds. Do not panic for I/O failures, missing users, or validation errors. `recover` is for isolating bugs in goroutines or plugin boundaries, not for normal control flow.

## Channel Patterns

Channels are for **communication and synchronization**, not as a general-purpose queue abstraction unless you have measured need.

### Buffered vs unbuffered

- **Unbuffered:** strong handoff semantics; sender and receiver rendezvous. Use when you need backpressure or clear "commit" points.
- **Buffered:** decouple producer and consumer rates up to capacity. Mis-sized buffers hide deadlocks and memory growth; size from design or profiling, not magic numbers.

### Direction in APIs

Expose **receive-only** (`<-chan T`) or **send-only** (`chan<- T`) types in function signatures when possible. It documents usage and prevents misuse.

### Closing

Only the **sender** should close a channel, and only when no more values will be sent. Closing a closed channel panics. Receivers use `for range` or `select` with `ok` to detect close.

### Fan-in / fan-out

- **Fan-out:** multiple workers read from one channel (often with a `WaitGroup` or `errgroup`).
- **Fan-in:** merge multiple channels into one with a dedicated goroutine per source; use `select` or a merge helper. Always ensure goroutines exit when inputs are done (context cancel, close, or explicit done channel).

### `select` defaults

- Use `default` in `select` for **non-blocking** sends/receives when appropriate.
- Without `default`, `select` blocks until one case is ready—prefer this for coordination to avoid busy loops.

## Package Design

Packages are the **unit of compilation and API stability**. Names and layout matter as much as types.

### Naming

- **Short, lowercase package names:** `http`, `json`, `user`, not `userServicePackage`.
- Avoid stutter: `user.User` is acceptable; prefer `user.Profile` over `user.UserProfile` when the package name already provides context.
- **Internal packages** (`internal/`) hide implementation from importers outside the module tree—use for large modules.

### What belongs in a package

- Cohesive functionality and shared invariants. Split when packages accrete unrelated concerns or cyclic imports appear.
- **Init order is fragile**—see anti-patterns below. Prefer explicit `New`/`Init` functions and dependency injection.

### Exported API surface

Export only what callers need. Unexported helpers keep flexibility for refactors. Document package behavior in **package comments** (the `// Package foo ...` block above `package foo`).

## `go vet`, `gofmt`, and mechanical correctness

### `gofmt`

Run `gofmt` (or configure the editor to format on save). **Go has one canonical formatting style**—debate belongs in design, not indentation.

### `go vet`

`go vet` catches common mistakes: printf format mismatches, unreachable code, suspicious locks, struct tags, and more. Run it in CI alongside tests.

### Static analysis in CI

Teams commonly add `staticcheck`, `golangci-lint`, or `govulncheck`. Treat linter output as part of the definition of "green" the same way tests are.

### Generated code

Keep generated files clearly marked (`// Code generated ... DO NOT EDIT.`) and exclude or segregate them in lint config when appropriate—never use that as an excuse to skip review on handwritten code.

## Table-Driven Tests

Prefer **one test function** with a slice of cases over many nearly identical `TestXxx` functions.

- Each case is a row: inputs, expected output, and sometimes a name.
- Use `t.Run(name, func(t *testing.T) { ... })` for isolation and clearer failure output.
- Use `t.Parallel()` **inside** subtests when cases do not share mutable state.
- Shared maps/slices mutated by tests must not run in parallel without synchronization—or give each subtest its own data.

See `examples/table-driven-test.md` for a full worked example.

## Concurrency: Context, WaitGroup, errgroup

### `context.Context`

- Pass `context.Context` as the **first parameter** of functions that do I/O or long work.
- Use `context.WithTimeout` / `WithDeadline` for bounded operations; use `WithCancel` for parent-driven shutdown.
- **Do not store contexts in structs** except for types that are request-scoped bridges (e.g. RPC handlers); prefer passing through call chains.
- Document whether a function respects cancellation (most should).

### `sync.WaitGroup`

Use a `WaitGroup` to wait for a **fixed set** of goroutines to finish. Pair every `Add(1)` with a `Done()` (typically via `defer wg.Done()`). Wait after all work is started, or structure `Add` before spawning to avoid races.

### `golang.org/x/sync/errgroup`

Use `errgroup.Group` (or the context-aware variant) when you need **wait + first error**. It combines cancellation propagation with waiting for a dynamic set of tasks—cleaner than manual channel plumbing for many patterns.

### Mutexes vs channels

**Share memory by communicating** is guidance, not dogma. Use `sync.Mutex` / `RWMutex` for protecting shared state; use channels for signaling and ownership transfer. Pick the construct that makes invariants obvious.

## Common Anti-Patterns

### `init()` abuse

`init()` runs in an undefined order across packages, complicates testing, and hides side effects. Prefer explicit `Register` calls, `main`-time setup, or lazy initialization with `sync.Once`.

### Interface pollution

Large interfaces with many methods force implementers to stub unused methods and make testing heavy. Prefer **small interfaces** (sometimes one method) composed at use sites.

### Naked returns

Named result parameters with **naked `return`** make refactors error-prone and hurt readability. Use named returns when they document meaning (e.g. `(n int, err error)` in low-level parsers) or defer modifies results—not to save typing.

### Panic for expected conditions

Missing files, bad user input, and downstream `500` responses are **normal errors**. Return `error`. Panic when the program cannot establish its own invariants (e.g. duplicate registration in a global map that should be impossible if the code is correct).

### Shadowing `err`

Repeated `if err := ...; err != nil` in the same scope is idiomatic. Accidental shadowing or reusing `err` after a failed branch causes subtle bugs. Use short variable declaration deliberately and consider `err` the most important name not to reuse carelessly.

### Global mutable state

Package-level mutable variables complicate concurrency and tests. Prefer dependency injection, `sync.Once`, or explicit registries owned by `main`.

## When in Doubt

Prefer **clear, boring code** that a new teammate can reason about in one pass. Reach for the standard library and well-trodden patterns (`context`, table tests, `%w`, small interfaces) before custom frameworks.

For deeper examples, see:

- `examples/table-driven-test.md`
- `examples/error-wrapping.md`
