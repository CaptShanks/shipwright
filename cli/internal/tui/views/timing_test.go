package views

import (
	"testing"
	"time"

	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
)

func TestFetchMarketplaceTiming(t *testing.T) {
	start := time.Now()
	msg := fetchMarketplace()
	elapsed := time.Since(start)

	t.Logf("fetchMarketplace completed in %v", elapsed)

	fetched, ok := msg.(common.MarketplaceFetchedMsg)
	if !ok {
		t.Fatalf("wrong message type: %T", msg)
	}
	if fetched.Err != nil {
		t.Logf("ERROR from fetch: %v", fetched.Err)
		t.Logf("This could cause the TUI to show 'Loading marketplace...' indefinitely")
		t.Logf("if the error message is not being routed properly")
	} else {
		t.Logf("Success: %d plugins loaded", len(fetched.Marketplace.Plugins))
	}

	if elapsed > 5*time.Second {
		t.Logf("WARNING: fetch took >5s - this would cause the TUI to appear stuck")
	}
}

func TestMarketplaceFullLifecycle(t *testing.T) {
	m := NewMarketplace()

	// Verify initial state
	view := m.View()
	if m.loading != true {
		t.Fatal("expected loading=true initially")
	}
	t.Logf("Initial view:\n%s", view)

	// Simulate the fetch completing
	msg := fetchMarketplace()
	m, _ = m.Update(msg)

	// Verify loaded state
	if m.loading {
		t.Fatal("still loading after receiving fetch result")
	}
	if !m.loaded {
		t.Fatal("not marked loaded after receiving fetch result")
	}

	view = m.View()
	t.Logf("After fetch view:\n%s", view)
}
