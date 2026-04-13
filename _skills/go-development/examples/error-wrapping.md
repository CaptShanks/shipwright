# Error Handling in Go: Wrapping, Sentinels, and Custom Types

Go errors are **values**. Production code uses a small set of patterns so callers can:

- show useful messages to operators and users,
- attach causality (`A failed because B failed`),
- branch on specific conditions with `errors.Is` and `errors.As`.

This document shows **complete examples** and **when to use each pattern**.

## 1. Simple errors with `errors.New`

Use for **static messages** that do not wrap another error and do not need structured fields.

```go
var ErrNotFound = errors.New("not found")
```

- Export the variable (`Err...`) when **callers** must compare with `errors.Is`.
- Keep the message lowercase (no punctuation at the end) unless you have a style guide saying otherwise—messages are often chained.

```go
func FindUser(ctx context.Context, id string) (*User, error) {
    u, ok := store.Get(id)
    if !ok {
        return nil, ErrNotFound
    }
    return u, nil
}
```

**When to use:** stable, API-level outcomes (not found, conflict, forbidden) where no underlying error exists.

## 2. Dynamic errors with `fmt.Errorf` (no wrap)

Use `%v` when the underlying error is **diagnostic only** and you **do not** want callers matching on it with `errors.Is`.

```go
if err != nil {
    return fmt.Errorf("parse config file %q: %v", path, err)
}
```

**When to use:** internal boundaries where the cause is logged/displayed but not part of your public contract. Prefer this over leaking driver-specific errors from your package’s documented API.

## 3. Wrapping with `fmt.Errorf` and `%w`

Use `%w` to attach context while preserving an **unwrap chain** for `errors.Is` / `errors.As`.

```go
func LoadProfile(ctx context.Context, userID string) (*Profile, error) {
    p, err := db.FetchProfile(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("load profile %q: %w", userID, err)
    }
    return p, nil
}
```

Caller:

```go
p, err := LoadProfile(ctx, id)
if errors.Is(err, ErrNotFound) {
    // handle missing profile
}
```

**When to use:** almost all error returns **up the stack** inside your module, whenever callers might need to distinguish causes or you use structured logging that walks `Unwrap()`.

**Rule:** one `%w` per `fmt.Errorf`—wrapping multiple errors requires `fmt.Errorf` with multiple `%w` (Go 1.20+) or a custom type implementing `Unwrap() []error`.

## 4. Sentinel errors

Sentinels are package-level `var` errors compared with `errors.Is` (not `==` on every path, because wrapped errors break `==`).

```go
package auth

import "errors"

var (
    ErrInvalidToken = errors.New("invalid token")
    ErrExpiredToken = errors.New("expired token")
)
```

Handler:

```go
claims, err := auth.ParseToken(token)
if err != nil {
    switch {
    case errors.Is(err, auth.ErrExpiredToken):
        return http.StatusUnauthorized, "token expired"
    case errors.Is(err, auth.ErrInvalidToken):
        return http.StatusUnauthorized, "invalid token"
    default:
        return http.StatusInternalServerError, "internal error"
    }
}
```

**When to use:**

- The condition is **part of your public API** and many callers will branch on it.
- The error carries **no extra data** beyond presence/absence.

**Avoid:** dozens of sentinels—prefer a small set or typed errors (below) when you need fields.

## 5. Custom error types and `errors.As`

Define a type when callers need **structured fields** (HTTP status, key, offset) or a **category** of errors with behavior.

```go
package quota

import "fmt"

type ExceededError struct {
    TenantID string
    Limit      int64
    Used       int64
}

func (e *ExceededError) Error() string {
    return fmt.Sprintf("quota exceeded for tenant %s: used %d of %d", e.TenantID, e.Used, e.Limit)
}
```

Prefer pointer receivers for `Error()` if the type holds slices or you want `errors.As` to match consistently on `*ExceededError`.

Caller:

```go
err := EnforceQuota(ctx, tenant)
var qe *quota.ExceededError
if errors.As(err, &qe) {
    log.Info("quota denial", "tenant", qe.TenantID, "limit", qe.Limit)
    return http.StatusTooManyRequests
}
```

**When to use:**

- Errors that naturally carry **domain data** operators or handlers need.
- You want `errors.As` without string matching.

Implement `Unwrap() error` on the type if the error **wraps** a lower-level cause:

```go
func (e *ExceededError) Unwrap() error { return e.Cause }
```

## 6. Wrapping across package boundaries

### Inside your module

Wrap liberally with context: **who** was doing **what** with **which inputs**.

```go
return fmt.Errorf("save order %s for user %s: %w", orderID, userID, err)
```

### At the public API edge of your package

1. **Document** which errors are stable (`ErrNotFound`, `*QuotaError`, etc.).
2. **Do not** expose every SQL or RPC string to callers unless that is intentional.
3. Wrap internal failures as **your** errors when the distinction matters:

```go
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
    b, err := c.inner.Get(ctx, key)
    if err != nil {
        if errors.Is(err, inner.ErrKeyNotFound) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("get %q: %w", key, err)
    }
    return b, nil
}
```

Here `ErrNotFound` is **your** sentinel; the wrapped `inner` error might still be inspectable with `errors.Is` if you used `%w`—decide whether that is desirable. Many teams **map** vendor errors to local sentinels at the boundary and use `%v` for the rest to avoid leaking implementation.

### Returning errors from `main`

Log the full chain at the top level; printing `%+v` with `github.com/pkg/errors`-style stack traces is optional (stdlib favors `slog` / `zap` with structured fields). For CLI tools:

```go
if err := run(os.Args[1:]); err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
}
```

## 7. Choosing a pattern (summary)

| Situation | Pattern |
|-----------|---------|
| Fixed, API-stable outcome | `var ErrX = errors.New(...)` + return `ErrX` |
| Add context, preserve unwrap | `fmt.Errorf("...: %w", err)` |
| Add context, **hide** unwrap from `Is`/`As` | `fmt.Errorf("...: %v", err)` or map to your sentinel |
| Need structured fields | Custom type + `errors.As` |
| Several related failures with fields | One type + enum / field, or small set of types |
| Assertion / impossible state | `panic` in development; prefer `error` in libraries |

## 8. Complete mini-example tying patterns together

```go
package document

import (
    "errors"
    "fmt"
    "io/fs"
)

// Stable errors for callers.
var ErrNotFound = errors.New("document not found")

// Detailed error for callers that need the path; still matches ErrNotFound via Unwrap.
type NotFoundError struct {
    Path string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("document %s: not found", e.Path)
}

func (e *NotFoundError) Unwrap() error {
    return ErrNotFound
}

func Open(path string, fsys fs.FS) ([]byte, error) {
    b, err := fs.ReadFile(fsys, path)
    if err != nil {
        if errors.Is(err, fs.ErrNotExist) {
            return nil, &NotFoundError{Path: path}
        }
        return nil, fmt.Errorf("read document %q: %w", path, err)
    }
    return b, nil
}
```

Consumer can use either:

```go
_, err := document.Open("x.txt", fsys)
if errors.Is(err, document.ErrNotFound) { /* ... */ }

var nf *document.NotFoundError
if errors.As(err, &nf) { /* use nf.Path */ }
```

In real code you would pick **one** primary style per error case to avoid redundancy—this illustrates that `Is` and `As` compose when types and sentinels are designed intentionally.

## 9. Anti-patterns (quick reference)

- **`panic` for I/O or validation** — return `error`.
- **Comparing with `==`** to exported third-party errors — use `errors.Is`.
- **String contains** (`strings.Contains(err.Error(), "not found")`) — brittle; use `Is`/`As` or stable error codes.
- **Logging every layer** — log at boundaries; return wrapped errors elsewhere.
- **Wrapping with `%w` without intent** — if the wrapped error is not part of your API contract, `%v` or mapping may be clearer.

Use `errors.Join` (Go 1.20+) when you genuinely accumulate **multiple independent errors** (e.g. validation of many fields) and need to preserve them all in one value.
