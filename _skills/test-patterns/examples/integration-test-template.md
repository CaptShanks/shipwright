# Integration test template (Go examples, language-agnostic ideas)

Integration tests verify **multiple real components together**: your code, a real database, HTTP handlers against real TCP, message consumers with a broker test instance, etc. The goal is **confidence in wiring and I/O**, not maximal speed.

## What “integration” means here

- **Includes real infrastructure** that unit tests fake or skip.
- **Slower** than unit tests; keep them **focused** on boundaries you own.
- **Isolated from production** and from **other parallel tests** (separate DB schema, random ports, unique topic names).

---

## Package layout and build tags (Go convention)

Keep integration tests **separate** so `go test ./...` stays fast for default workflows.

**Option A — `_test` package in same folder:** good for black-box testing `mypkg` as `mypkg_test`.

**Option B — `//go:build integration`:** gate expensive suites.

```go
//go:build integration

package billing_test

import "testing"

func TestInvoiceRepository_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	// ...
}
```

Run with:

```bash
go test -tags=integration -count=1 ./...
```

**Language-agnostic:** use **tags, env gates, or separate CI jobs** so developers can run quick suites locally and push full integration runs to CI or opt-in commands.

---

## External dependency setup and teardown

**Pattern: start once per package, isolate per test.**

- **Package-level:** expensive process/container startup via `sync.Once` in `TestMain` (amortize cost).
- **Test-level:** **truncate tables**, **reset state**, or **new schema per test** depending on isolation needs.

```go
//go:build integration

package billing_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

var pgDSN string // set in TestMain (testcontainers, docker compose, or CI service)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Typical sources, pick one for your repo:
	// 1) CI injects DATABASE_URL (GitHub Actions service container, CircleCI, etc.)
	if dsn := os.Getenv("INTEGRATION_POSTGRES_DSN"); dsn != "" {
		pgDSN = dsn
		os.Exit(m.Run())
	}

	// 2) Local/dev: spin Postgres via testcontainers (or docker compose helper)
	dsn, cleanup, err := startPostgresContainer(ctx)
	if err != nil {
		// Fail fast with an actionable message (CI should set INTEGRATION_POSTGRES_DSN or start containers).
		fmt.Fprintln(os.Stderr, "integration postgres:", err)
		os.Exit(1)
	}
	defer cleanup()

	pgDSN = dsn
	os.Exit(m.Run())
}

// startPostgresContainer is a stand-in for testcontainers-go or a wrapper around
// `docker run`. Real implementations return a DSN with sslmode appropriate for tests.
func startPostgresContainer(ctx context.Context) (dsn string, cleanup func(), err error) {
	_ = ctx
	// Example shape with testcontainers (API names vary by version):
	//
	// req := testcontainers.ContainerRequest{
	// 	Image:        "postgres:16-alpine",
	// 	ExposedPorts: []string{"5432/tcp"},
	// 	Env: map[string]string{
	// 		"POSTGRES_USER":     "test",
	// 		"POSTGRES_PASSWORD": "test",
	// 		"POSTGRES_DB":       "integration",
	// 	},
	// 	WaitingFor: wait.ForLog("database system is ready to accept connections"),
	// }
	// c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
	// 	ContainerRequest: req,
	// 	Started:          true,
	// })
	// if err != nil { return "", nil, err }
	// host, err := c.Host(ctx)
	// if err != nil { _ = c.Terminate(ctx); return "", nil, err }
	// port, err := c.MappedPort(ctx, "5432")
	// if err != nil { _ = c.Terminate(ctx); return "", nil, err }
	// dsn = fmt.Sprintf("postgres://test:test@%s:%s/integration?sslmode=disable", host, port.Port())
	// cleanup = func() { _ = c.Terminate(context.Background()) }
	// return dsn, cleanup, nil

	return "", nil, errors.New("set INTEGRATION_POSTGRES_DSN or implement startPostgresContainer")
}
```

**Teardown responsibilities:**

- stop containers or child processes
- close connection pools
- remove temp directories

Register cleanup **where resources are created** so failures mid-setup still release partial resources.

---

## Testcontainers (Go sketch)

[Testcontainers](https://golang.testcontainers.org/) (and analogs in other languages) spin **real** databases, brokers, and cloud emulators in CI-friendly ways.

**Concepts:**

- **wait strategies:** don’t assert “ready” too early; wait for logs, TCP port, or a successful ping.
- **reuse (optional):** some modes reuse containers across tests on a dev machine; still keep **data isolated**.
- **Ryuk / reaper:** understand how your library cleans orphaned containers when tests crash.

```go
func startPostgres(ctx context.Context) (dsn string, cleanup func(), err error) {
	// Pseudocode-ish: API varies by module version
	// req := testcontainers.ContainerRequest{ Image: "postgres:16", ExposedPorts: ..., Env: ... }
	// c, err := testcontainers.GenericContainer(ctx, ...)
	// host, err := c.Host(ctx)
	// port, err := c.MappedPort(ctx, "5432")
	// dsn = fmt.Sprintf("postgres://user:pass@%s:%s/db?sslmode=disable", host, port.Port())
	// cleanup = func() { _ = c.Terminate(ctx) }
	return "", func() {}, nil
}
```

**Language-agnostic takeaway:** integration tests should **wait for readiness**, **retry connection** with backoff, and **always terminate** infrastructure.

---

## Test isolation strategies

Pick one consistent strategy per suite:

| Strategy | Pros | Cons |
|----------|------|------|
| **Fresh database per test** | strongest isolation | slower |
| **Shared DB, transaction rollback per test** | fast for SQL | doesn’t cover commit semantics; not all drivers support nested tx the same way |
| **Shared DB, truncate between tests** | simple | easy to miss tables; FK order matters |
| **Unique tenant/key prefix per test** | good for shared services | requires disciplined schema design |

**Rule:** two tests running **in parallel** must not observe each other’s data. If you cannot guarantee that, **forbid parallel integration tests** (`t.Parallel` off) or **shard by schema**.

---

## Database fixtures

**Prefer factories** over static SQL blobs:

- each test states **only what it cares about**
- schema evolution doesn’t break every `.sql` file

```go
type seedInvoice struct {
	ID     string
	Amount int64
	Status string
}

func insertInvoice(ctx context.Context, t *testing.T, db *sql.DB, inv seedInvoice) {
	t.Helper()
	_, err := db.ExecContext(ctx, `
		INSERT INTO invoices (id, amount_cents, status)
		VALUES ($1, $2, $3)
	`, inv.ID, inv.Amount, inv.Status)
	if err != nil {
		t.Fatalf("seed invoice: %v", err)
	}
}
```

**Migrations:** run migrations **once** at suite startup against the container DB; tests assume **current schema**.

**Cleanup:** for integration tests that **commit**, delete by **primary key** or **test run ID**:

```go
runID := uuid.NewString()
t.Cleanup(func() { deleteTestData(context.Background(), t, db, runID) })
```

---

## HTTP integration: `httptest` + real server

Spin the **real** router/listener, not the handler in isolation, when middleware matters.

```go
func TestAPI_CreateInvoice(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	srv, db := newTestServer(t) // starts *httptest.Server, wires deps
	t.Cleanup(srv.Close)

	body := strings.NewReader(`{"customer_id":"c1","amount_cents":1000}`)
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/invoices", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d", res.StatusCode)
	}

	// Assert DB side effects with a fresh query / sqlmock-less approach
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM invoices WHERE customer_id = $1`, "c1").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("invoice rows = %d", count)
	}
}
```

**Language-agnostic:** hit **loopback** with real HTTP; assert **status, headers, body schema**, and **persistent state** if the endpoint commits.

---

## Concurrency in integration tests

If your app uses **worker pools**, **outbox processors**, or **listeners**, integration tests may need:

- a **synchronization barrier** (poll DB until row reaches terminal state, with timeout)
- **context cancellation** to stop background goroutines in `t.Cleanup`

Avoid unbounded `time.Sleep`; prefer **eventual consistency** checks:

```go
waitFor(t, 5*time.Second, func() bool {
	var status string
	_ = db.QueryRow(`SELECT status FROM invoices WHERE id = $1`, id).Scan(&status)
	return status == "PAID"
})
```

---

## Cleanup checklist (integration)

- [ ] Containers/processes **terminated** on success and failure paths
- [ ] **No fixed ports**; allocate `0` or dynamic mapping
- [ ] **Unique resource names** (topics, buckets, tenants) per test or per run
- [ ] **No reliance on execution order**
- [ ] **Short mode** or build tags to skip in tight loops
- [ ] **-count=1** in CI to disable cache for flaky-suite diagnostics

---

## When this template is the wrong tool

- Pure logic with no I/O → **unit tests**
- Full browser journey → **E2E** (Playwright, Selenium) with fewer cases
- Third-party systems without emulators → **contract tests** + **recorded fixtures** + **toggled live smoke**

Use integration tests where **your code meets reality**—schema, wire format, and side effects—and invest in **isolation** so the suite stays **trustworthy** as it grows.
