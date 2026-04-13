# Failure Mode Analysis Checklist

For every component or integration, ask these questions:

## Dependency Failures

- [ ] **What if the database is unreachable?** -- Does the application crash, hang, or degrade gracefully?
- [ ] **What if an API returns 500?** -- Is the error retried? With backoff? Is there a fallback?
- [ ] **What if DNS resolution fails?** -- Are resolved addresses cached? Is there a fallback?
- [ ] **What if the network is slow (not down)?** -- Are timeouts configured? Will they cascade?
- [ ] **What if a dependency returns invalid data?** -- Is the response validated before use?

## Resource Exhaustion

- [ ] **What if memory is constrained?** -- Are large payloads streamed or buffered entirely?
- [ ] **What if disk space runs out?** -- Do writes fail cleanly or corrupt data?
- [ ] **What if connections are exhausted?** -- Is there a connection pool with a bounded size?
- [ ] **What if goroutines/threads leak?** -- Do all goroutines have exit conditions?
- [ ] **What if file descriptors are exhausted?** -- Are files and sockets closed promptly?

## Load and Concurrency

- [ ] **What happens under 10x expected load?** -- Linear degradation or catastrophic failure?
- [ ] **What if two requests modify the same resource simultaneously?** -- Race condition? Lost update?
- [ ] **What if a batch job takes longer than the interval?** -- Overlapping runs? Data corruption?
- [ ] **What if the input is 100x larger than expected?** -- Timeout? OOM? Pagination?

## State and Data

- [ ] **What if the process crashes mid-operation?** -- Is state consistent? Can it resume?
- [ ] **What if a message is delivered twice?** -- Is processing idempotent?
- [ ] **What if events arrive out of order?** -- Does the system handle reordering?
- [ ] **What if a cache becomes stale?** -- Is there a TTL? Can the system work without the cache?
