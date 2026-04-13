# Error Handling Patterns Checklist

## Every Error Path Should

- [ ] **Be handled explicitly** -- no ignored error returns
- [ ] **Add context** -- wrap with what operation was being attempted
- [ ] **Preserve the cause** -- don't discard the original error when wrapping
- [ ] **Clean up resources** -- close files, connections, and channels on error paths
- [ ] **Fail fast** -- return early rather than continuing in a degraded state

## Wrapping Errors

Add context at each abstraction boundary. The final error message should read like a chain of operations:

```
"failed to create user: failed to insert into database: connection refused"
```

Each layer adds what it was trying to do, and the root cause is preserved.

## Error Classification

Classify errors by what the caller should do:

| Category | Caller Action | Example |
|----------|---------------|---------|
| Retryable | Retry with backoff | Network timeout, rate limit |
| Permanent | Stop trying, report to user | Invalid input, not found |
| Bug | Alert engineering | Nil pointer, index out of bounds |
| Infrastructure | Circuit-break and degrade | Database down, service unavailable |

## Logging vs Returning

Choose one at each layer:
- **Leaf functions** (lowest level): return the error, let the caller decide
- **Boundary functions** (API handler, main): log the error with full context and return a safe response
- **Never both** at the same layer -- it creates duplicate log entries

## Sentinel Errors and Error Types

Use sentinel errors or typed errors when callers need to distinguish error kinds:
- Sentinel: `var ErrNotFound = errors.New("not found")` -- check with `errors.Is`
- Typed: `type ValidationError struct{...}` -- check with `errors.As`
- Avoid comparing error strings -- they're brittle and break on message changes
