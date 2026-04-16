# Migration Readiness Checklist

## Principle: Migrate from Strength, Not Desperation

An architectural migration succeeds when the current architecture is well-understood, boundaries already exist (even if imperfect), and the team has capacity beyond keeping the lights on. Migrating while drowning is how you get a distributed monolith.

## 1. Understand the Current State

- [ ] Complete the [Pattern Assessment](pattern-assessment.md) checklist — you cannot migrate from something you have not accurately identified
- [ ] Document the **actual** dependency graph, not the intended one
- [ ] Identify all shared data stores, shared libraries, and cross-cutting concerns
- [ ] Map which teams own which components and where ownership is ambiguous
- [ ] Catalog all integration points: APIs, events, shared databases, file drops, cron jobs

## 2. Clarify the Target State

- [ ] Define the target architecture pattern with concrete structural rules, not just a name
- [ ] Specify what "done" looks like — what properties must the system exhibit?
- [ ] Identify which bounded contexts or service boundaries exist in the target
- [ ] Document which constraints the target architecture enforces (deployment independence, data ownership, interface contracts)
- [ ] Verify the target pattern actually solves the problems driving the migration (not just industry fashion)

## 3. Validate the Business Case

- [ ] Articulate the specific pain the migration solves (deployment coupling, team scaling, performance, compliance)
- [ ] Quantify the cost of not migrating (incident frequency, developer hours lost, blocked feature delivery)
- [ ] Estimate the migration timeline in months, not weeks — migrations always take longer than expected
- [ ] Confirm leadership commitment to sustained investment (migrations abandoned halfway are worse than not starting)
- [ ] Identify the first measurable milestone that demonstrates value (not "phase 1 complete" but "team X can deploy independently")

## 4. Assess Technical Prerequisites

- [ ] Is there adequate test coverage to catch regressions during restructuring? (target: critical paths > 70%)
- [ ] Is CI/CD in place and reliable? (you will need fast feedback loops)
- [ ] Can you deploy frequently? (migrations are safest as many small changes, not one big bang)
- [ ] Is observability sufficient to detect problems in the migrated system? (logging, metrics, tracing)
- [ ] Are there integration or contract tests between components that will become separate?

```
# Assess test coverage before starting
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total

# Check for existing integration tests
find . -name '*_integration_test*' -o -name '*_e2e_test*' | wc -l
```

## 5. Evaluate Data Separation Readiness

- [ ] Identify all shared database tables and which components read/write each
- [ ] Determine if shared tables can be split without data duplication or can tolerate eventual consistency
- [ ] Plan the data migration strategy: dual-write, change data capture, event sourcing, or big-bang cutover
- [ ] Assess whether foreign key constraints span future service boundaries
- [ ] Verify that no component uses database-level joins across future boundary lines for critical-path operations

## 6. Assess Team Readiness

- [ ] Does the team have experience operating the target architecture? (microservices require distributed systems skills)
- [ ] Is there bandwidth beyond feature delivery to sustain migration work?
- [ ] Are on-call and incident response processes ready for the operational complexity of the target?
- [ ] Has the team agreed on shared conventions for the target (API contracts, error handling, observability standards)?
- [ ] Is there a designated migration lead or working group with authority to make cross-team decisions?

## 7. Plan the Migration Strategy

- [ ] Choose an approach: **strangler fig** (incremental replacement), **branch by abstraction** (swap implementations behind interfaces), or **parallel run** (run both, compare)
- [ ] Identify the first extraction candidate — pick the component with the fewest inbound dependencies and clearest boundary
- [ ] Define rollback criteria — what signals trigger reverting a migration step?
- [ ] Plan for a coexistence period where old and new architectures run simultaneously
- [ ] Establish contract tests between migrated and non-migrated components

### Strangler fig sequence

```
1. Identify a vertical slice (one business capability end-to-end)
2. Build the new implementation behind the same interface
3. Route traffic to the new implementation (feature flag, load balancer, DNS)
4. Validate with production traffic (shadow mode or canary)
5. Decommission the old implementation
6. Repeat for the next slice
```

## 8. Infrastructure Readiness

- [ ] Can the infrastructure support the target topology? (separate databases, message brokers, service mesh)
- [ ] Are networking and security controls ready for inter-service communication?
- [ ] Is there capacity for running old and new systems in parallel during transition?
- [ ] Are deployment pipelines ready for the target's deployment model (per-service pipelines, independent releases)?
- [ ] Is secrets management ready for the increased number of service identities?

## 9. Risk Assessment

- [ ] Identify the highest-risk migration step (usually data separation or contract changes)
- [ ] Plan a spike or proof-of-concept for the riskiest step before committing
- [ ] Document what happens if the migration is abandoned partway — can you operate the hybrid state indefinitely?
- [ ] Identify external dependencies that constrain the migration timeline (vendor contracts, compliance deadlines, API consumers)
- [ ] Assess the blast radius of each migration step — what breaks if it goes wrong?

## Go/No-Go Decision

Before starting, confirm all of these:

- [ ] Current architecture is accurately documented (not assumed)
- [ ] Target architecture solves a real, quantified problem
- [ ] Test coverage and CI/CD are sufficient for safe refactoring
- [ ] Team has the skills and bandwidth for sustained migration work
- [ ] A rollback or coexistence plan exists for every migration step
- [ ] Leadership is committed to the timeline and trade-offs

## Anti-Patterns

- **Big bang rewrite** — Stopping feature development to rebuild from scratch. This almost always fails, delivers no value during the rewrite, and the new system misses undocumented behaviors from the old one.
- **Resume-driven architecture** — Migrating to microservices because the team wants to learn Kubernetes, not because the system needs it.
- **Migrating without tests** — Restructuring code that has no test coverage. You will introduce regressions and not know until production.
- **Ignoring the data problem** — Extracting services while leaving them all connected to the same database. You have not migrated; you have added network latency.
- **Declaring victory at the halfway point** — Celebrating the first extracted service while 80% of the system remains in the monolith. Budget for the full journey.
