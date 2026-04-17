package views

import (
	"testing"

	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
)

func TestFetchMarketplace(t *testing.T) {
	msg := fetchMarketplace()

	fetched, ok := msg.(common.MarketplaceFetchedMsg)
	if !ok {
		t.Fatalf("expected MarketplaceFetchedMsg, got %T", msg)
	}

	if fetched.Err != nil {
		t.Fatalf("fetch error: %v", fetched.Err)
	}

	if fetched.Marketplace == nil {
		t.Fatal("marketplace is nil")
	}

	t.Logf("Fetched %d plugins", len(fetched.Marketplace.Plugins))
	for _, p := range fetched.Marketplace.Plugins {
		t.Logf("  %s [%s] v%s", p.Name, p.Category, p.Version)
	}
}

func TestMarketplaceUpdateReceivesMsg(t *testing.T) {
	m := NewMarketplace()

	mp := &registry.Marketplace{
		Plugins: []registry.MarketplaceItem{
			{Name: "test-agent", Category: "agents", Version: "1.0.0", Description: "A test"},
		},
	}

	msg := common.MarketplaceFetchedMsg{Marketplace: mp}
	m, _ = m.Update(msg)

	if !m.loaded {
		t.Fatal("marketplace not marked as loaded after receiving MarketplaceFetchedMsg")
	}
	if m.loading {
		t.Fatal("marketplace still loading after receiving MarketplaceFetchedMsg")
	}
	if len(m.allItems) != 1 {
		t.Fatalf("expected 1 item, got %d", len(m.allItems))
	}
	if m.err != nil {
		t.Fatalf("unexpected error: %v", m.err)
	}
}
