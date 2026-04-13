# Unit test template (Go examples, language-agnostic ideas)

This document encodes a **canonical unit test shape** and Go idioms for **table-driven tests**, **subtests**, **parallelism**, and **cleanup**. The principles apply broadly: isolate the unit, control inputs, assert observable outcomes, and leave no shared mutable state behind.

## Arrange / Act / Assert (AAA)

Structure each test so readers see **setup**, **exercise**, and **verification** at a glance. In Go, you rarely need comment banners—**blank lines** between phases are enough.

**Good pattern:**

```go
func TestDiscount_AppliesPercentage(t *testing.T) {
	t.Parallel()

	// Arrange
	p := Product{SKU: "A", UnitPriceCents: 1000}
	line := LineItem{Product: p, Qty: 2}

	// Act
	total, err := ApplyPercentDiscount(line, 10) // 10% off

	// Assert
	if err != nil {
		t.Fatalf("ApplyPercentDiscount: %v", err)
	}
	if got, want := total, 1800; got != want {
		t.Fatalf("total = %d, want %d", got, want)
	}
}
```

**Why it helps:** failures point to **wrong setup** vs **wrong behavior** vs **wrong expectation**. Mixed phases obscure which broke.

---

## Before: sloppy unit test

Problems illustrated:

- asserts **implementation** (internal helper calls) instead of **outcome**
- **shared mutable state** (`lastSKU`) → order-dependent flakiness
- no subtest boundaries → one failure hides others
- **sleep** instead of synchronization
- missing cleanup for temp resources

```go
var lastSKU string // BAD: package-level test state

func TestOrderStuff(t *testing.T) {
	lastSKU = "X"
	o := &Order{}
	o.AddLine(LineItem{SKU: "X", Qty: 1})
	if lastSKU != "X" {
		t.Fatal("wrong sku") // BAD: testing side effect, not price/totals
	}
	time.Sleep(10 * time.Millisecond) // BAD: flaky timing

	// Mutates global lastSKU elsewhere in package — next test may break
	o.AddLine(LineItem{SKU: "Y", Qty: 2})
	if o.lines != 2 { // BAD: asserts private/internal representation
		t.Fatal("lines bad")
	}
}
```

---

## After: disciplined unit test

Improvements:

- **only public behavior** asserted (totals, errors)
- **no shared state** between tests
- **table-driven** cases with **subtests**
- **`t.Parallel()`** where safe
- **`t.Cleanup`** for resources

```go
func TestOrder_AddLine_Total(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		lines      []LineItem
		wantTotal  int64
		wantErr    bool
	}{
		{
			name: "single line",
			lines: []LineItem{
				{SKU: "A", UnitPriceCents: 500, Qty: 2},
			},
			wantTotal: 1000,
		},
		{
			name: "multiple lines",
			lines: []LineItem{
				{SKU: "A", UnitPriceCents: 300, Qty: 1},
				{SKU: "B", UnitPriceCents: 700, Qty: 1},
			},
			wantTotal: 1000,
		},
		{
			name: "reject zero qty",
			lines: []LineItem{
				{SKU: "A", UnitPriceCents: 100, Qty: 0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable for parallel subtests
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: fresh subject per subtest
			order := NewOrder()

			// Act
			var err error
			for _, li := range tt.lines {
				err = order.AddLine(li)
				if err != nil {
					break
				}
			}

			// Assert
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got := order.TotalCents(); got != tt.wantTotal {
				t.Fatalf("TotalCents = %d, want %d", got, tt.wantTotal)
			}
		})
	}
}
```

**Language-agnostic takeaway:** enumerate cases as **data**, run each case in **isolation**, and assert **contracts** (return values, errors, observable state), not private fields.

---

## Table-driven tests and subtests

**Use tables when:**

- multiple inputs map to expected outputs
- the **logic under test is the same** across rows
- you want reviewers to spot **missing cases** quickly

**Use subtests (`t.Run`) when:**

- you need **parallelism per row**
- you want **granular failure reporting** in CI logs
- setup differs slightly per case (optional `before` func in the table)

**Parallelism footgun in Go:** always capture the loop variable:

```go
for _, tt := range tests {
	tt := tt // required when t.Parallel inside loop
	t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
		// ...
	})
}
```

If you cannot run cases in parallel (shared heavyweight setup), **omit `t.Parallel`** on those tests or **move expensive setup** to `TestMain` / package-level sync.Once with **immutable** fixtures.

---

## Cleanup: `t.Cleanup` and `defer`

Prefer **`t.Cleanup`** for test-scoped resources so **early returns and subtests** still release resources. It runs in **reverse registration order** after the test function returns.

```go
func TestWriteManifest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // Go 1.15+: auto cleanup, preferred for filesystem

	f, err := os.Create(filepath.Join(dir, "out.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })

	// Act: write under dir ...
}
```

**When `defer` is fine:** purely local scope in a single function with no subtests registering more cleanup—still correct, but **`t.Cleanup` scales better** as tests grow helpers.

**Language-agnostic:** pair **resource acquisition** with **release** at the same abstraction level; never rely on “the next test will overwrite global state.”

---

## What to assert in unit tests

Assert **public API contracts**:

- return values and errors
- observable mutations on the receiver / returned aggregate
- calls to **explicit ports** (interfaces) when interaction is part of the contract—prefer **fakes** recording calls over brittle mock frameworks when possible

Avoid asserting:

- private fields unless refactoring legacy code with no public seam
- **call order** unless ordering is a **specified** guarantee
- **exact error string text**—prefer **typed errors** or **stable sentinel values** (`errors.Is` / `errors.As` in Go)

---

## Quick checklist

- [ ] AAA visible; no mixed phases
- [ ] No shared mutable package/global state
- [ ] Table + `t.Run` for multiple scenarios
- [ ] `t.Parallel` only where isolated
- [ ] Loop variable captured when parallel
- [ ] `t.TempDir` / `t.Cleanup` (or equivalent) for resources
- [ ] Assertions on **behavior**, not **internals**
