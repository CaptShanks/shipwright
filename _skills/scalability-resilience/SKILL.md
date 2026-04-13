---
name: scalability-resilience
description: >-
  Single points of failure, graceful degradation, idempotency, resource
  management, and resilience patterns. Use when designing or reviewing systems
  that involve I/O, external dependencies, concurrency, retries, or resource
  lifecycle management.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---

# Scalability and Resilience

This skill covers the engineering thinking required to build systems that work reliably under real-world conditions -- where networks are unreliable, resources are finite, dependencies fail, and load is unpredictable.

## Core Principles

### Assume Failure
Every external dependency will eventually fail: databases, APIs, file systems, DNS, the network itself. Design for when, not if. The question is never "will it fail?" but "what happens when it fails?"

### Design for Recovery
A system that can recover from failure quickly is more valuable than one that rarely fails but takes hours to recover. Prioritize:
1. Detection (know it's broken)
2. Isolation (limit the blast radius)
3. Recovery (get back to working state)
4. Prevention (stop it from happening again)

### Idempotency
Operations should be safe to retry. If a request times out and gets retried, the second execution should produce the same result as the first -- not duplicate data, double-charge a customer, or send a second notification.

### Backpressure
When a system is overwhelmed, it should push back rather than collapse. Unbounded queues, unlimited goroutines, and "accept everything" patterns lead to cascading failures.

## What to Always Consider

1. **What happens if this dependency is down?** -- timeout, fallback, or graceful error?
2. **What happens if this runs twice?** -- idempotent or will it cause duplicates?
3. **What happens under 10x load?** -- linear scaling, exponential resource use, or hard limits?
4. **What happens if this takes 100x longer than expected?** -- timeouts, cancellation, deadline propagation?
5. **What resources does this hold?** -- connections, file handles, memory, goroutines?
6. **What is the blast radius if this fails?** -- isolated to one user, one feature, or the whole system?

## Resource Lifecycle

Every resource that is acquired must be released:
- File handles: close after reading/writing
- Network connections: close or return to pool
- Database connections: return to pool, don't leak
- Goroutines/threads: ensure they terminate, avoid leaking
- Locks: unlock in the same function, preferably via defer
- Timers: stop timers that are no longer needed
- Channels: close when no more values will be sent
- Temporary files: clean up on exit, even on error paths

## Patterns

### Timeouts on All I/O
Every network call, database query, and external process invocation needs a timeout. Without one, a hung dependency silently blocks resources until the system runs out.

### Retry with Exponential Backoff
When retrying failed operations:
- Start with a short delay (100-500ms)
- Double the delay on each retry
- Add random jitter to prevent thundering herd
- Set a maximum number of retries
- Only retry operations that are idempotent and the error is transient

### Circuit Breaker
When a dependency is consistently failing:
- After N consecutive failures, stop trying (circuit open)
- Periodically attempt a single request (circuit half-open)
- On success, resume normal operation (circuit closed)
- Return a cached response or degraded result while the circuit is open

### Graceful Degradation
When a non-critical feature fails, the core experience should continue:
- Recommendation engine is down → show popular items instead
- Avatar service is slow → show initials placeholder
- Analytics endpoint is unreachable → buffer locally and retry later
