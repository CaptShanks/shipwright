# Code Review Checklist

Use this checklist as a structured pass over a pull request. Not every item applies to every change—skip irrelevant sections and go deeper on high-risk areas (auth, persistence, parsers, concurrency, public APIs).

---

## Naming and readability

- [ ] Names (variables, functions, types, files) reflect **intent** and domain vocabulary, not implementation accidents.
- [ ] Booleans read as assertions (`isLoaded`, `hasAccess`); functions that perform actions use **verbs** (`fetch`, `validate`, `persist`).
- [ ] Names are **consistent** with surrounding code and existing public APIs; no redundant context (`User.userName` in a `User` type).
- [ ] Abbreviations are **team-standard** or widely understood; obscure abbreviations are avoided or documented.
- [ ] Magic numbers and strings are **named constants** or enums when they carry meaning beyond the immediate line.
- [ ] Comments explain **why** (constraints, invariants, non-obvious tradeoffs), not **what** the code already states.
- [ ] User-facing copy (errors, UI, API messages) is **clear, accurate, and appropriately toned**.

---

## Error handling and robustness

- [ ] Errors are **detected**: no swallowed exceptions, ignored return values, or empty `catch` blocks without justification.
- [ ] Errors are **classified**: retryable vs. fatal, client vs. server, validation vs. system—where the distinction matters.
- [ ] Error paths are **reachable and tested** (not only the happy path).
- [ ] Wrapped or chained errors **preserve cause** and add **actionable context** (operation, resource id), without leaking secrets.
- [ ] Resource cleanup happens on **success and failure** (connections, files, locks, goroutines, subscriptions).
- [ ] Timeouts, cancellation, and **backpressure** are handled for I/O and cross-service calls where appropriate.
- [ ] Idempotency and **partial-failure** behavior are defined for operations that can be retried or replayed.
- [ ] Invariants are **enforced or documented**; defensive checks exist at trust boundaries.

---

## Tests and test coverage

- [ ] New behavior has **automated tests** at the right level (unit, integration, contract) per team policy.
- [ ] Tests **fail for the right reason**: assertions target behavior, not implementation trivia that will churn.
- [ ] Edge cases are covered: **empty, max, malformed, boundary** values; off-by-one; null/optional cases.
- [ ] Regressions for fixed bugs include a **test that would have failed before the fix**.
- [ ] Flaky patterns are avoided: real time without control, unordered collections assumed ordered, shared mutable state.
- [ ] Test data is **minimal and readable**; factories/helpers are used consistently.
- [ ] External systems are **mocked, stubbed, or containerized** per team norms—not accidentally hitting production.
- [ ] Coverage gaps in **critical paths** (auth, billing, migrations) are explicitly acknowledged if intentionally deferred.

---

## Complexity and control flow

- [ ] Control flow is **easy to follow**: early returns reduce nesting; deep nesting is justified or refactored.
- [ ] Cyclomatic complexity is **appropriate**; long functions are split where they mix unrelated concerns.
- [ ] Loops and recursion have **clear termination** and bounded work where needed.
- [ ] Concurrency is **correct**: no data races, deadlocks, or unsafe use of shared mutable state; locks/scopes are minimal.
- [ ] Async code handles **ordering and failure** explicitly (promises, futures, callbacks, actors).
- [ ] Feature flags and conditional compilation do not leave **unreachable or untested** combinations without a plan.

---

## DRY (Don’t Repeat Yourself)

- [ ] Duplication is **intentional or eliminated**: copy-paste with drift is flagged; “rule of three” or team standard applied for extraction.
- [ ] Shared abstractions **encode one concept**; accidental similarity is not forced into a premature abstraction.
- [ ] Configuration and defaults are **single-sourced** where multiple components must stay aligned.
- [ ] Repeated error-handling or validation patterns use **shared helpers** consistent with existing modules.

---

## SOLID and modular design

- [ ] **Single responsibility**: modules, types, and functions have one clear reason to change.
- [ ] **Open/closed**: extension points (interfaces, strategies, plugins) are used where new variants are expected.
- [ ] **Liskov**: subtypes honor contracts; mocks and implementations are substitutable without surprises.
- [ ] **Interface segregation**: consumers depend on **small, focused** surfaces, not “god” interfaces.
- [ ] **Dependency inversion**: high-level logic depends on **abstractions**; construction/wiring lives at the edge.
- [ ] Coupling is **directional**: no cycles; dependencies flow toward stable domain/core packages.
- [ ] Visibility is **minimal** (private by default); public surface is deliberate.

---

## API design (libraries, services, events)

- [ ] Public API is **minimal, coherent, and hard to misuse** (types enforce valid states where possible).
- [ ] Naming and semantics match **industry or team conventions** (HTTP verbs, status codes, RPC style, event names).
- [ ] Request/response models are **versioned or evolvable** where consumers are external or long-lived.
- [ ] Defaults are **safe**; dangerous operations require explicit parameters or flags.
- [ ] Pagination, filtering, and sorting behaviors are **defined and stable** for list endpoints.
- [ ] Errors returned to clients are **consistent** (codes, shapes) and free of sensitive internals.
- [ ] Idempotency keys, deduplication, or **exactly-once** expectations are documented where relevant.

---

## Backwards compatibility and migrations

- [ ] **Breaking changes** are identified: API fields, event schemas, file formats, config keys, database columns.
- [ ] Deprecations follow **policy**: timeline, migration path, logging/metrics for old usage.
- [ ] Database migrations are **safe to deploy** (expand/contract, backfills, locking, rollback considerations).
- [ ] **Two-phase deploys** or feature flags are used when producer and consumer cannot update atomically.
- [ ] Default behavior preserves **existing users’ expectations** unless an intentional break is approved and communicated.
- [ ] Version bumps (semver, package, API version header) match the **actual compatibility** story.

---

## Documentation and operability

- [ ] User-facing docs (**README**, runbooks, OpenAPI, changelog) are updated when behavior or setup changes.
- [ ] **Non-obvious operational behavior** is documented: queues, retries, rate limits, cache TTLs, feature flags.
- [ ] Observability is adequate: **logs, metrics, traces** at useful points; log levels are appropriate.
- [ ] Dashboards/alerts are updated or tickets filed when **SLO-relevant** behavior changes.
- [ ] Config and env vars are **documented** with defaults, required vs. optional, and examples.

---

## Performance and scalability

- [ ] Hot paths avoid **unnecessary allocation**, N+1 queries, or repeated expensive work.
- [ ] Asymptotic complexity is **appropriate** for expected data sizes (sorting, nested loops on large collections).
- [ ] Caching has **correct invalidation** or TTL strategy; stale data risk is understood.
- [ ] Payload sizes and **streaming** are considered for large responses or uploads.
- [ ] Resource limits (memory, connections, thread pools) are **bounded** under load.
- [ ] Database indexes and query shapes support **expected access patterns**.

---

## Security surface

- [ ] **Inputs** from users, partners, files, and networks are validated (type, size, format, allowlists).
- [ ] **Authn** is enforced where required; sessions/tokens are validated correctly.
- [ ] **Authz** is checked at the **resource** level, not only route or UI level.
- [ ] **Injection** risks addressed: SQL, command, LDAP, template, header, log forging—use safe APIs and encoding.
- [ ] **Secrets** are not in code, logs, URLs, or error messages; secret rotation and storage follow team standards.
- [ ] **Crypto** uses vetted libraries; no custom ciphers; randomness is from secure sources where required.
- [ ] **Deserialization** of untrusted data is safe (avoid unsafe parsers or gadget chains).
- [ ] **SSRF, open redirects, path traversal** considered for URLs and filesystem access.
- [ ] **CORS, CSP, cookies** (HttpOnly, Secure, SameSite) are correct for the threat model.
- [ ] **Dependencies** are pinned/scanned; known-vulnerable versions are not introduced without mitigation.
- [ ] **PII** handling complies with retention, minimization, and logging redaction expectations.

---

## Meta (process and completeness)

- [ ] PR description states **purpose**, **scope**, and **how to verify** the change.
- [ ] Linked issues/tickets and **release notes** needs are satisfied where applicable.
- [ ] Generated files and **lockfiles** are intentional; spurious churn is absent.
- [ ] CI checks relevant to the change are **green** or failures are explained and acceptable.
