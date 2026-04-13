# Resource Management Checklist

## Acquisition and Release

- [ ] **Every opened resource has a corresponding close** -- file, connection, lock, channel
- [ ] **Cleanup is deferred immediately after acquisition** -- `defer file.Close()` right after `os.Open()`
- [ ] **Error paths clean up too** -- resources opened before an error are still released
- [ ] **Cleanup order is correct** -- LIFO (last opened, first closed) via defer stack

## Connections

- [ ] **Connection pools are bounded** -- max connections configured, not unlimited
- [ ] **Pool health checks are enabled** -- stale connections are detected and removed
- [ ] **Connections are returned to pool, not leaked** -- every borrow has a return
- [ ] **Connection timeouts are set** -- dial timeout, read timeout, idle timeout

## Timeouts

- [ ] **Every network call has a timeout** -- HTTP client, database query, gRPC call
- [ ] **Timeouts cascade correctly** -- outer timeout > inner timeout
- [ ] **Context cancellation is propagated** -- child operations respect parent deadlines
- [ ] **Timeouts are reasonable** -- not too short (premature failure) or too long (resource holding)

## Concurrency

- [ ] **Goroutines/threads have bounded lifetimes** -- they terminate when work is done
- [ ] **Worker pools are bounded** -- semaphore or buffered channel limits concurrency
- [ ] **Shared state is protected** -- mutex, atomic, or channel-based synchronization
- [ ] **No goroutine leaks** -- every started goroutine has a clear exit path via context or done channel

## Retry Budget

- [ ] **Maximum retries are configured** -- not infinite
- [ ] **Backoff increases between retries** -- exponential with jitter
- [ ] **Only retryable errors trigger retries** -- not validation failures or 4xx responses
- [ ] **Retry budget is shared** -- total retries across the call chain don't compound
