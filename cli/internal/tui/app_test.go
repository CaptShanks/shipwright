package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
)

func TestAppLazyLoadsMarketplace(t *testing.T) {
	app := NewApp("test")

	// Simulate window size
	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app = model.(App)

	// App starts on Dashboard
	if app.activeView != ViewDashboard {
		t.Fatalf("expected Dashboard, got view %d", app.activeView)
	}

	// Marketplace should not be loaded yet
	view := app.marketplace.View()
	if !strings.Contains(view, "Loading marketplace") {
		t.Fatal("marketplace should show loading before first visit")
	}

	// Switch to marketplace — this triggers the fetch
	cmd := app.setView(ViewMarketplace)
	if cmd == nil {
		t.Fatal("setView(ViewMarketplace) should return an Init cmd on first visit")
	}

	// Simulate the fetch completing (message arrives while marketplace IS active)
	mp := &registry.Marketplace{
		Plugins: []registry.MarketplaceItem{
			{Name: "test-agent", Category: "agents", Version: "1.0.0", Description: "A test agent"},
			{Name: "test-mcp", Category: "mcps", Version: "1.0.0", Description: "A test MCP"},
		},
	}
	model, _ = app.Update(common.MarketplaceFetchedMsg{Marketplace: mp})
	app = model.(App)

	// Marketplace should now show data
	view = app.marketplace.View()
	t.Logf("Marketplace view:\n%s", view)

	if strings.Contains(view, "Loading marketplace") {
		t.Fatal("marketplace still shows loading after data was received")
	}

	// Second visit should NOT re-trigger fetch
	app.setView(ViewDashboard)
	cmd = app.setView(ViewMarketplace)
	if cmd != nil {
		t.Fatal("setView(ViewMarketplace) should NOT return cmd on subsequent visits")
	}
}
