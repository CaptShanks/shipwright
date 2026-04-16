---
name: architecture-patterns
description: >-
  How to identify architectural styles from code evidence, distinguish practiced
  architecture from aspirational structure, and recognize common failure modes.
  Use when analyzing codebases, designing solutions, or reviewing systems for
  structural integrity.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---

# Architecture Patterns

This skill is about **recognizing architecture from code**, not about knowing pattern names from textbooks. Directory names are aspirations. README claims are hypotheses. The code is the experiment. Your job is to determine what architecture is actually practiced—which may differ substantially from what was intended.

## How to Identify Architecture

Architecture is not what you call it. It is the set of **hard-to-reverse decisions** that shape how the system behaves at runtime: what components exist, how they communicate, where state lives, and which boundaries are enforced versus decorative.

### Evidence over labels

A `domain/` package does not mean domain-driven design. A `microservices/` directory does not mean the system is decomposed into independent services. An `internal/` folder does not mean encapsulation is enforced.

To identify the actual architecture, trace:

1. **Entry points** — Where does execution begin? (`main()`, HTTP handler registration, event consumer setup, CLI parser)
2. **Dependency direction** — Who imports whom? Does domain logic depend on infrastructure, or does infrastructure serve domain logic?
3. **Data flow** — How does data move from input to storage to output? Synchronous call chain? Event bus? Shared database?
4. **Deployment units** — What gets deployed together? Separate binaries, containers, or a single artifact?
5. **Failure boundaries** — When one component fails, what else breaks? If everything breaks, you have a monolith regardless of package count.

### The gap between intended and practiced architecture

Every codebase has architectural intent (what the team wanted) and architectural reality (what the code does). Common divergences:

- **Layered intent, spaghetti reality** — Clean package names but imports that skip layers, direct database access from HTTP handlers, and business logic scattered across controllers.
- **Microservices intent, distributed monolith reality** — Separate repos and deployments, but every request fans out to five services synchronously, shared database schemas, and no team can deploy independently.
- **Clean architecture intent, anemic domain reality** — Domain types exist but are plain data holders with no behavior. All logic lives in "services" that operate on domain objects externally.
- **Event-driven intent, request-response reality** — An event bus exists but events trigger synchronous handlers that block on downstream calls, making the system effectively synchronous with extra latency.

Report the gap, not just the reality. Understanding *what was attempted* helps explain the codebase's evolution and predicts where the next structural problem will emerge.

## Pattern Recognition Guide

### Layered Architecture

**What it looks like in code:**
- Packages or modules named by technical concern: `handlers/`, `services/`, `repositories/`, `models/`
- Dependency flows top-down: handlers → services → repositories → database
- Each layer has a clear role: presentation, business logic, data access
- Types are often passed between layers with transformation at boundaries (DTOs, domain models, database entities)

**Signs it is practiced well:**
- No layer skips another (handlers never call repositories directly)
- Business logic is in the service layer, not in handlers or database queries
- Each layer can be tested independently with mocks at the boundary below

**Signs it has degraded:**
- Handlers contain business logic ("fat controllers")
- Services are thin pass-throughs that add no value
- Database queries encode business rules (complex WHERE clauses that should be domain logic)
- Models are shared across layers without transformation (tight coupling through shared types)

**Common failure mode:** Layers become a bureaucracy. Every new feature requires touching three files that do trivially different things (handler → service → repo) with no real separation of concerns—just separation of files.

### Hexagonal / Ports and Adapters

**What it looks like in code:**
- A `core/` or `domain/` package with zero infrastructure imports (no database drivers, HTTP frameworks, or cloud SDKs)
- Interfaces defined inside the domain that infrastructure implements (`type UserRepository interface`, `type NotificationSender interface`)
- Adapter packages that implement domain interfaces: `postgres/`, `http/`, `grpc/`, `inmemory/`
- Application services that orchestrate domain logic through ports (interfaces)

**Signs it is practiced well:**
- The domain package compiles without any infrastructure dependency
- Swapping an adapter (Postgres → DynamoDB) requires no domain code changes
- Tests for domain logic use in-memory fakes, not real databases
- The dependency arrow points inward: infrastructure depends on domain, never the reverse

**Signs it has degraded:**
- Domain types import ORM tags (`gorm:`, `bson:`) or serialization annotations
- "Ports" are 1:1 mirrors of a specific database's query API (not abstracting, just wrapping)
- The domain package has grown a dependency on a framework because "it was easier"
- Adapters contain business logic that should be in the domain

**Common failure mode:** Over-abstraction. Ten interfaces with one implementation each, adapter layers that add indirection without enabling substitution, and a domain so "pure" it cannot express anything useful without calling out to infrastructure through a maze of ports.

### Microservices

**What it looks like in code:**
- Separate deployment units (separate `main()` functions, Dockerfiles, or CI pipelines)
- Inter-service communication via HTTP/gRPC calls, message queues, or event streams
- Each service owns its data store (no shared databases)
- Service discovery or routing configuration (load balancers, service mesh, DNS)
- Independent versioning and release cycles

**Signs it is practiced well:**
- Teams can deploy their service independently without coordinating with other teams
- Each service has a bounded context with clear ownership
- Inter-service contracts are versioned and tested (contract tests, schema registries)
- Failure in one service degrades gracefully rather than cascading

**Signs it is actually a distributed monolith:**
- Deploying service A requires deploying service B at the same time
- Shared database schemas across services (one migration can break another service)
- Synchronous call chains where every request traverses 3+ services serially
- A "shared library" that contains business logic and must be version-locked across services
- No service can handle a request if any downstream service is unavailable

**Common failure mode:** Premature decomposition. Splitting a system into services before understanding domain boundaries creates services that are too chatty, too coupled, and harder to operate than the monolith they replaced—with the added cost of network serialization, distributed tracing, and eventual consistency.

### Event-Driven / Message-Based

**What it looks like in code:**
- Message broker or event bus infrastructure (Kafka, RabbitMQ, SQS, NATS, or in-process event dispatchers)
- Publishers that emit events without knowing who consumes them
- Subscribers/handlers registered for specific event types
- Event schemas or types defined as data contracts
- Idempotency logic in handlers (handling duplicate delivery)

**Signs it is practiced well:**
- Publishers and consumers are decoupled—adding a new consumer requires no publisher changes
- Events carry enough context for consumers to act without calling back to the publisher
- Dead letter queues and retry logic handle processing failures
- Event schemas are versioned and backward compatible
- The system remains functional when a consumer is temporarily offline (events queue)

**Signs it has degraded:**
- Events are used as RPC (publish, then immediately poll for a response)
- Event handlers call back to the publisher synchronously, creating circular dependencies
- No dead letter queue—failed events are silently dropped
- Events carry only IDs, forcing consumers to make synchronous calls to resolve state
- No ordering guarantees where ordering matters (parallel consumers reorder events)

**Common failure mode:** Event soup. So many event types with unclear ownership that nobody can trace a business operation end-to-end. Debugging requires correlating logs across a dozen consumers, and adding a new feature means understanding which events trigger which handlers in which order.

### Modular Monolith

**What it looks like in code:**
- Single deployment unit (one binary, one container) with internal module boundaries
- Modules communicate through well-defined internal APIs (function calls, internal event bus, or interface contracts)
- Each module owns its database tables or schema partition
- Shared infrastructure (HTTP server, connection pool) but isolated business logic

**Signs it is practiced well:**
- Module boundaries are enforced (internal packages, visibility rules, linter checks for cross-module imports)
- Each module can be understood, tested, and modified without deep knowledge of other modules
- Extracting a module into a separate service would require only wiring changes, not redesign
- Shared code is genuinely shared (utilities, auth middleware), not domain logic leaked across modules

**Signs it has degraded into a big ball of mud:**
- Any module can import any other module's internal types
- Circular dependencies between modules
- A change in module A requires changes in modules B, C, and D
- No clear module ownership—features are spread across many packages with no coherent grouping
- "Shared" packages contain business logic specific to one module but imported by convenience

**Common failure mode:** Module boundaries start strong and erode under deadline pressure. A "temporary" cross-module import becomes permanent. Shared types accumulate fields from multiple modules. Within a year, the monolith is modular in name only.

## Cross-Pattern Assessment

When analyzing a codebase, assess these dimensions regardless of the architectural style:

### Coupling

How much does changing one component require changing others? High coupling means high coordination cost.

- **Tight coupling signals:** shared mutable state, direct database access across modules, chain of synchronous calls, shared types that accumulate fields from multiple concerns.
- **Loose coupling signals:** communication through interfaces, events, or well-defined APIs. Changes to internals don't ripple outward.

### Cohesion

How well does each component's contents relate to a single purpose? Low cohesion means the component is a grab bag.

- **High cohesion signals:** every function in a package serves the same concept. The package can be described in one sentence.
- **Low cohesion signals:** a package named `utils`, `common`, or `helpers` with unrelated functions. A service that handles both user auth and billing because they share a database table.

### Boundary enforcement

Are architectural boundaries enforced by tooling, or only by convention?

- **Enforced:** compiler visibility rules (`internal/` in Go), module systems, linter rules that flag cross-boundary imports, separate repositories.
- **Convention only:** comments saying "do not import this directly," README warnings, verbal agreements. Convention boundaries erode under pressure.

### Data ownership

Who owns each piece of data? Shared data ownership is the most common source of architectural degradation.

- **Clear ownership:** one module writes a table, others read through its API. Schema changes are owned by one team.
- **Shared ownership:** multiple modules write to the same table. Schema changes require coordination across teams. This is the distributed monolith's defining trait.

### Evolutionary fitness

How well does the architecture support the changes the system needs to make? An architecture optimized for a startup's needs may be wrong for an enterprise's needs, and vice versa.

- **Evolving well:** new features fit naturally into existing boundaries. Teams can work independently. The architecture enables the organization's delivery cadence.
- **Evolving poorly:** every feature requires cross-cutting changes. Adding a simple field touches five services. The architecture fights the organization structure (inverse Conway).

## Common Smells Across All Patterns

- **God package/class** — One component that everything depends on and that knows about everything. Indicates collapsed boundaries.
- **Circular dependencies** — A depends on B depends on A. Always a boundary violation. Resolve by extracting shared concerns or inverting the dependency.
- **Shotgun surgery** — Adding a feature requires touching many unrelated files. Indicates missing abstractions or wrong decomposition.
- **Leaky abstraction** — Infrastructure concerns (SQL, HTTP headers, queue message formats) visible in business logic. Indicates missing boundary.
- **Connascence of timing** — Components that must execute in a specific undocumented order. Indicates implicit state machine without explicit representation.
- **Feature envy** — A function that mostly accesses data from another module. Suggests the function belongs in that other module.
- **Primitive obsession at boundaries** — Passing strings, maps, and raw bytes between components instead of typed contracts. Indicates missing domain types at the interface.
