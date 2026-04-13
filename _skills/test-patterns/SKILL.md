---
name: test-patterns
description: >-
  Testing strategies, mocking, fixture management, edge case identification, and
  the testing pyramid. Use when designing test suites, writing unit or
  integration tests, choosing what to mock, or hardening tests around boundaries,
  state, and concurrency.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---

# Test patterns and testing strategy

Use this skill when designing test suites, choosing what to mock, structuring fixtures, or hardening tests around boundaries, state, and concurrency. It favors **fast feedback**, **clear intent**, and **tests that survive refactors**.

## The testing pyramid

The pyramid is a **risk and cost** model, not a rigid ratio:

- **Unit tests (base, wide):** Exercise pure logic, small collaborators, and invariants with minimal I/O. They should be numerous, cheap, and deterministic.
- **Integration tests (middle):** Prove that **your code plus real boundaries** (database, message broker, HTTP client to a fake server, filesystem in a temp dir) behave correctly. Fewer than unit tests, slower, still automated.
- **End-to-end / UI / smoke (top, narrow):** Validate critical user journeys and deployment health. The fewest tests, the highest flake potential, the longest runtime.

**Anti-patterns:** Inverting the pyramid (mostly E2E), or treating “integration” as “call production.” **Healthy shape:** most assertions live in fast tests; the top validates what only the full stack can reveal.

## Tests as specification

Treat tests as **executable documentation** of behavior you intend to keep:

- Name tests after **observable outcomes** and **rules**, not implementation (`rejects_negative_quantity` not `calls_validate`).
- Prefer **one primary behavior per test** (or one table row) so failures pinpoint regressions.
- Document **invariants** explicitly: “balance never negative,” “idempotent retry,” “at-most-once side effect.”
- When requirements change, **update tests first** or alongside code so the suite encodes the new contract.

This mindset reduces “testing for coverage” and increases “testing for correctness.”

## Boundary value analysis

Bugs cluster at **edges** of domains:

- Numeric: `0`, `1`, `max-1`, `max`, `min`, negatives, overflow/underflow where relevant.
- Strings: empty, single character, Unicode normalization, very long, whitespace-only, delimiter-heavy.
- Collections: empty, one element, duplicates, sorted vs unsorted assumptions.
- Time: DST transitions, leap seconds (where APIs expose them), “now” at window boundaries.
- Enums / states: first, last, deprecated values, unknown future values (forward compatibility).

**Technique:** For each input dimension, list **representative classes** (valid interior, valid boundary, invalid boundary, invalid interior). Combine dimensions sparingly (pairwise / orthogonal arrays) when combinatorial explosion threatens.

## State machine testing

When behavior depends on **sequence** (sessions, workflows, connection pools, caches, sagas), model explicit **states** and **transitions**:

- Identify **states**, **events**, and **guarded transitions** (what is legal from where).
- Test **reachable paths**: happy path, cancellation, timeout, retry after partial failure.
- Test **illegal transitions**: ensure the system rejects or no-ops safely (no silent corruption).
- For stateful components under refactoring, keep a **transition table** (or table-driven tests) aligned with product rules.

**Property-style angle:** For finite state machines, random walks with invariants (e.g., “no double commit”) can complement hand-picked paths—especially for protocol-ish code.

## Concurrency testing

Concurrency defects are **non-deterministic**; tests should **narrow the schedule** or **amplify races**:

- **Determinism hooks:** injectable clocks, sequenced executors, barriers, or channels so you control ordering in tests.
- **Stress loops:** run a critical section thousands of times under `-race` (Go) or equivalent; pair with small, focused scenarios.
- **Model:** which variables are shared? Which locks protect them? Test **contended** and **uncontended** paths.
- **Timeouts:** avoid flaky sleeps; prefer synchronization primitives or polling with a bounded deadline in tests.

**Goal:** not “prove no bug exists,” but **catch classes of races** (lost updates, double close, read-after-free patterns) and encode expected memory/thread-safety contracts.

## When to mock vs use real dependencies

### Prefer real dependencies when

- The dependency is **fast, local, and deterministic** (in-memory repo, temp sqlite, `httptest.Server`).
- You care about **query correctness, serialization, or wire compatibility**—mocks lie about these.
- The risk is **integration failure**, not algorithmic error.

### Prefer mocks, fakes, or stubs when

- The real system is **slow, flaky, or unavailable** in CI (third-party SaaS, hardware).
- You need to simulate **faults** (timeouts, 500s, partial responses) reproducibly.
- You must assert **interaction contracts** (called once, idempotent retry) **without** standing up the whole service.

### Guidelines

- **Mock outward, test inward:** at module boundaries, substitute **ports** (interfaces), not deep internals.
- **Fakes > mocks** when behavior matters: an in-memory fake queue that preserves ordering teaches more than `EXPECT().Times(3)`.
- **Avoid mocking what you own** unless isolating a pure function from a side-effecting shell—often a sign the API needs splitting.
- **Verify fewer interactions:** over-specified mocks couple tests to implementation and break on refactors.

### Fixture management

- **Fresh vs shared:** default to **isolation** (each test gets clean state). Share expensive setup only when cost dominates and **immutability** is guaranteed.
- **Factories, not mystery blobs:** builders/factories for test data beat copy-pasted JSON; expose only fields tests care about.
- **Deterministic IDs:** seed RNGs or use fixed UUIDs when snapshots or ordering matter.
- **Cleanup:** `t.Cleanup`, `defer`, or container teardown—prefer **register cleanup where resources are created** so subtests and early returns stay safe.

## Edge case identification (practical checklist)

Ask:

1. **What can be nil, empty, or missing?**
2. **What are the numeric and temporal boundaries?**
3. **What happens on duplicate, replay, or retry?**
4. **What happens under partial failure (before commit, after commit, message redelivery)?**
5. **What happens if the caller violates ordering assumptions?**
6. **What are security-relevant edges** (authorization, injection, oversize payloads)?

Capture these as named scenarios in **table-driven tests** or **explicit subtests** so missing cases are visible in review.

## Closing heuristic

**Unit tests** prove your logic and invariants.**Integration tests** prove your wiring and real I/O assumptions.**E2E tests** prove the product still works for humans.**Mocks** are scaffolding—use them to control fault and time, not to avoid thinking about reality.
