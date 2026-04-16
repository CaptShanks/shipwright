# Architectural Pattern Assessment Checklist

## Principle: Read the Code, Not the Labels

Directory names are aspirations. README claims are hypotheses. Identify the architecture that is **practiced**, not the one that was intended.

## 1. Map Entry Points

- [ ] Identify all execution entry points (`main()`, HTTP handler registration, CLI parser, event consumer setup)
- [ ] Document which frameworks or libraries control the request lifecycle (Express, Gin, Spring, Bubble Tea, etc.)
- [ ] Check for multiple entry points that indicate separate deployment units vs. a single monolithic binary
- [ ] Note any background workers, cron jobs, or scheduled tasks that run independently

## 2. Trace Dependency Direction

- [ ] Pick 3-5 core domain operations and trace the import/dependency chain from entry point to storage
- [ ] Determine whether domain logic depends on infrastructure, or infrastructure serves domain logic
- [ ] Check for circular dependencies between packages/modules (always indicates a boundary violation)
- [ ] Map which packages have zero external dependencies (candidates for domain core)

```
# Quick dependency check in Go
go list -f '{{.ImportPath}} -> {{join .Imports "\n  "}}' ./...

# In Python, use import-linter or pydeps
import-linter --config .importlinter

# In TypeScript, use madge
npx madge --circular src/
```

## 3. Assess Layer Structure

- [ ] Are packages organized by **technical concern** (handlers, services, repos) or by **domain** (users, orders, billing)?
- [ ] Do handlers contain business logic? (fat controllers = degraded layering)
- [ ] Are service layers thin pass-throughs that add no value? (bureaucratic layers)
- [ ] Do database queries encode business rules via complex WHERE clauses?
- [ ] Are types shared across layers without transformation? (tight coupling through shared models)

## 4. Evaluate Boundary Enforcement

- [ ] Are boundaries enforced by the compiler/runtime (Go `internal/`, Java module system, TypeScript project references)?
- [ ] Or are boundaries convention-only (comments, READMEs, verbal agreements)?
- [ ] Do linter rules or CI checks flag cross-boundary imports?
- [ ] Has any boundary been violated by "temporary" workarounds that became permanent?

## 5. Assess Coupling

- [ ] Can you change one component's internals without modifying others?
- [ ] Is there shared mutable state between components?
- [ ] Do components communicate through interfaces/contracts or through direct references?
- [ ] Count the fan-out: how many other components does the busiest component depend on?
- [ ] Identify any "god package" that everything imports (collapsed boundary)

## 6. Assess Cohesion

- [ ] Can each package/module be described in one sentence?
- [ ] Are there packages named `utils`, `common`, `helpers`, or `shared` with unrelated functions?
- [ ] Does each package serve a single domain concept or technical concern?
- [ ] Look for "feature envy" — functions that mostly access data from another module

## 7. Evaluate Data Ownership

- [ ] Does each module/service own its data exclusively (single writer)?
- [ ] Are there shared database tables written by multiple components?
- [ ] Can one team change a schema without coordinating with other teams?
- [ ] Is data shared via APIs/events, or via direct database access?

## 8. Check Deployment Topology

- [ ] How many independently deployable artifacts exist (binaries, containers, serverless functions)?
- [ ] Can any component be deployed without redeploying others?
- [ ] Do separate "services" share a database, message format, or deployment pipeline?
- [ ] Is there a shared library containing business logic that must be version-locked across services?

## 9. Identify the Practiced Pattern

Based on evidence gathered above, classify the architecture:

- [ ] **Layered** — packages by technical concern, top-down dependency flow, layer-skipping indicates degradation
- [ ] **Hexagonal / Ports & Adapters** — domain core with zero infrastructure imports, interfaces defined in domain, adapters implement ports
- [ ] **Microservices** — separate deployment units, independent data stores, inter-service communication via network
- [ ] **Modular Monolith** — single deployment unit with enforced internal module boundaries
- [ ] **Event-Driven** — message broker infrastructure, publishers decoupled from consumers, idempotent handlers
- [ ] **Big Ball of Mud** — no discernible boundaries, everything depends on everything

## 10. Document the Gap

- [ ] State the **intended** architecture (from README, docs, team description)
- [ ] State the **practiced** architecture (from code evidence above)
- [ ] Describe specific divergences with file/package references
- [ ] Identify the top 3 forces that caused the divergence (deadline pressure, missing tooling, team turnover)

## Anti-Patterns

- **Label shopping** — Calling a codebase "hexagonal" because it has a `domain/` folder, despite domain types importing ORM annotations.
- **Architecture by aspiration** — Drawing a clean diagram on a whiteboard but never enforcing the boundaries in code.
- **Pattern matching by directory name** — Concluding "microservices" because there are multiple directories, even though they deploy as one binary.
- **Ignoring runtime behavior** — Analyzing only static structure while the system's real architecture is defined by synchronous call chains, shared caches, and deployment coupling.
