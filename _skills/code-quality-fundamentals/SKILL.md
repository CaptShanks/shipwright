---
name: code-quality-fundamentals
description: >-
  Naming conventions, SOLID principles, error handling patterns, function design,
  and code organization fundamentals. Use when writing new code, refactoring
  existing code, or reviewing pull requests for maintainability and readability.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---

# Code Quality Fundamentals

This skill covers the engineering principles that separate professional code from code that merely works. Quality code is readable, maintainable, testable, and predictable.

## Naming

Names are the most important documentation. A well-named function needs no comment.

### Variables
- Name should describe **what it holds**, not how it's used: `userEmail` not `inputStr`
- Booleans should read as assertions: `isValid`, `hasPermission`, `canRetry`
- Avoid single-letter names except in tight loops (`i`, `j`) or well-known conventions (`ctx`, `err`)
- Avoid abbreviations unless universally understood in the domain: `req`/`resp` are fine, `usrMgr` is not
- Collections should be plural: `users`, `orderItems`, `pendingTasks`

### Functions
- Name should describe **what it does**, starting with a verb: `validateInput`, `fetchUser`, `calculateTotal`
- Boolean-returning functions should read as questions: `isExpired()`, `hasAccess()`, `canProceed()`
- Avoid vague verbs: `handle`, `process`, `manage`, `do` -- what specifically does it do?
- Side-effect-free functions should have noun-like names: `userName()`, `totalPrice()`

### Types and Interfaces
- Types describe **what it is**: `User`, `OrderItem`, `ConnectionPool`
- Interfaces describe **what it can do**: `Reader`, `Validator`, `Notifier`
- Avoid `I` prefix or `Interface` suffix: `Reader` not `IReader` or `ReaderInterface`

## SOLID Principles (Applied, Not Academic)

### Single Responsibility
A function does one thing. A type manages one concept. If you need the word "and" to describe what something does, it likely needs splitting.

### Open/Closed
Prefer extending behavior through composition and interfaces rather than modifying existing code. Add new implementations, don't change existing ones.

### Liskov Substitution
Any implementation of an interface should be safely swappable for any other. If your `MockDatabase` behaves fundamentally differently from `PostgresDatabase`, the interface is wrong.

### Interface Segregation
Don't force consumers to depend on methods they don't use. Prefer small, focused interfaces over large ones. A `Reader` interface with one method is better than a `FileManager` with ten.

### Dependency Inversion
Depend on abstractions, not concrete implementations. Accept interfaces in function parameters. Inject dependencies rather than constructing them internally.

## Error Handling

### Principles
- Errors are values, not exceptions to be thrown and caught
- Every error should be handled explicitly or propagated with context
- The error message should answer: what failed, why, and with what input?
- Wrap errors at abstraction boundaries to add context without losing the original cause

### Anti-Patterns
- Silencing errors: `_ = doSomething()` -- never ignore an error return
- Bare returns: `return err` without adding context -- wrap with what operation failed
- Logging and returning: don't log an error AND return it -- choose one, or you get duplicate logs
- Stringly typed errors: use typed errors or sentinel values for errors that callers need to match on
- Panic for expected conditions: reserve panics for truly unrecoverable states

## Function Design

### Size
If a function exceeds 40-50 lines, it likely does too much. Extract helper functions with descriptive names.

### Parameters
- Prefer fewer parameters (0-3 ideal, 4+ is a code smell)
- Group related parameters into a struct/config object
- Avoid boolean parameters -- they make call sites unreadable: `createUser(true, false)` vs `createUser(opts)`
- Required parameters first, optional last

### Return Values
- Return errors explicitly rather than using sentinel values (e.g., `-1` for "not found")
- Return early on error -- avoid deep nesting
- Named return values only when they improve readability at the call site

## Code Organization

### DRY at the Right Level
Duplication is better than the wrong abstraction. Extract common code only when:
- The pattern has appeared 3+ times
- The duplicated logic serves the same conceptual purpose (not just happens to look similar)
- A change to one instance would require changing all others

### Cohesion
Group related functionality together. A package/module should have a clear theme. If you can't describe what a package does in one sentence, it may need splitting.

### Dependencies
- Import from specific packages, not umbrella packages
- Avoid circular dependencies -- they indicate poor module boundaries
- Minimize transitive dependencies -- each dependency is a liability
