package tui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CaptShanks/shipwright/cli/internal/installer"
	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
	"github.com/CaptShanks/shipwright/cli/internal/tui/components"
	"github.com/CaptShanks/shipwright/cli/internal/tui/views"
)

const (
	ViewDashboard = iota
	ViewMarketplace
	ViewInstalled
	ViewDetail
	ViewMcpConfig
)

var tabNames = []string{"Dashboard", "Marketplace", "Installed"}

type batchPhase int

const (
	batchNone batchPhase = iota
	batchPickTarget
	batchPickScope
	batchInstalling
)

type App struct {
	tabs        components.TabBar
	statusBar   components.StatusBar
	toast       components.Toast
	dashboard   views.Dashboard
	marketplace views.Marketplace
	installed   views.Installed
	detail      views.Detail
	mcpConfig   views.McpConfig
	activeView       int
	prevView         int
	marketplaceReady bool
	width       int
	height      int
	version     string
	ready       bool

	batchItems        []registry.MarketplaceItem
	batchPhase        batchPhase
	batchTargets      []string
	batchTargetChecked []bool
	batchTargetCursor int
	batchScopeLocal   bool
	batchSpinner      spinner.Model
}

func NewApp(version string) App {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return App{
		tabs:            components.NewTabBar(tabNames),
		toast:           components.NewToast(),
		dashboard:       views.NewDashboard(),
		marketplace:     views.NewMarketplace(),
		installed:       views.NewInstalled(),
		detail:          views.NewDetail(),
		mcpConfig:       views.NewMcpConfig(),
		version:         version,
		batchScopeLocal: true,
		batchSpinner:    sp,
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.dashboard.Init(),
		a.installed.Init(),
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.tabs.Width = msg.Width
		a.ready = true

	case tea.KeyMsg:
		if a.batchPhase != batchNone {
			return a.handleBatchKey(msg)
		}

		if a.activeView == ViewDetail || a.activeView == ViewMcpConfig {
			if msg.String() == "esc" {
				a.activeView = a.prevView
				a.tabs.SetActive(a.prevView)
				return a, nil
			}
		} else {
			switch {
			case key.Matches(msg, Keys.Quit):
				return a, tea.Quit
			case key.Matches(msg, Keys.Tab):
				a.tabs.Next()
				if cmd := a.setView(a.tabs.Active); cmd != nil {
					cmds = append(cmds, cmd)
				}
			case key.Matches(msg, Keys.ShiftTab):
				a.tabs.Prev()
				if cmd := a.setView(a.tabs.Active); cmd != nil {
					cmds = append(cmds, cmd)
				}
			case key.Matches(msg, Keys.Num1):
				if cmd := a.setView(ViewDashboard); cmd != nil {
					cmds = append(cmds, cmd)
				}
			case key.Matches(msg, Keys.Num2):
				if cmd := a.setView(ViewMarketplace); cmd != nil {
					cmds = append(cmds, cmd)
				}
			case key.Matches(msg, Keys.Num3):
				if cmd := a.setView(ViewInstalled); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case common.ToastExpiredMsg:
		a.toast.HandleExpiry(msg)

	case common.InstallCompleteMsg:
		if msg.Err != nil {
			cmds = append(cmds, a.toast.Show("Install failed: "+msg.Err.Error(), common.ToastError))
		} else {
			cmds = append(cmds, a.toast.Show(msg.Plugin+" installed", common.ToastSuccess))
		}

	case common.UninstallCompleteMsg:
		if msg.Err != nil {
			cmds = append(cmds, a.toast.Show("Uninstall failed: "+msg.Err.Error(), common.ToastError))
		} else {
			cmds = append(cmds, a.toast.Show(msg.Plugin+" removed", common.ToastSuccess))
		}

	case common.BatchInstallCompleteMsg:
		a.batchPhase = batchNone
		a.batchItems = nil
		n := len(msg.Succeeded)
		nFail := len(msg.Failed)
		if nFail == 0 {
			cmds = append(cmds, a.toast.Show(fmt.Sprintf("%d plugins installed", n), common.ToastSuccess))
		} else {
			cmds = append(cmds, a.toast.Show(fmt.Sprintf("%d installed, %d failed", n, nFail), common.ToastWarn))
		}
		cmds = append(cmds, a.installed.Init())

	case spinner.TickMsg:
		if a.batchPhase == batchInstalling {
			var cmd tea.Cmd
			a.batchSpinner, cmd = a.batchSpinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case common.NavigateMsg:
		a.prevView = a.activeView
		a.activeView = msg.View
		if msg.View == ViewMcpConfig {
			if row, ok := msg.Context.([]string); ok && len(row) >= 3 {
				cmds = append(cmds, a.mcpConfig.Load(row[0], row[1], row[2]))
			}
		}
	}

	sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height - 4}

	// Data messages must reach their target view regardless of which view is active.
	switch msg.(type) {
	case common.StateFetchedMsg:
		if a.activeView != ViewDashboard {
			var cmd tea.Cmd
			a.dashboard, cmd = a.dashboard.Update(msg)
			cmds = append(cmds, cmd)
		}
		if a.activeView != ViewInstalled {
			var cmd2 tea.Cmd
			a.installed, cmd2 = a.installed.Update(msg)
			cmds = append(cmds, cmd2)
		}
		if a.activeView != ViewDetail {
			var cmd3 tea.Cmd
			a.detail, cmd3 = a.detail.Update(msg)
			cmds = append(cmds, cmd3)
		}
	}

	switch a.activeView {
	case ViewDashboard:
		var cmd tea.Cmd
		a.dashboard, cmd = a.dashboard.Update(msg)
		a.dashboard, _ = a.dashboard.Update(sizeMsg)
		cmds = append(cmds, cmd)

	case ViewMarketplace:
		var cmd tea.Cmd
		a.marketplace, cmd = a.marketplace.Update(msg)
		a.marketplace, _ = a.marketplace.Update(sizeMsg)
		cmds = append(cmds, cmd)

		if len(a.marketplace.SelectedBatch) > 0 {
			a.batchItems = a.marketplace.SelectedBatch
			a.marketplace.ClearBatch()
			a.initBatchPicker()
		} else if a.marketplace.SelectedItem != nil {
			item := a.marketplace.SelectedItem
			a.marketplace.ClearSelection()
			a.prevView = ViewMarketplace
			a.activeView = ViewDetail
			cmds = append(cmds, a.detail.SetItem(item))
		}

	case ViewInstalled:
		var cmd tea.Cmd
		a.installed, cmd = a.installed.Update(msg)
		a.installed, _ = a.installed.Update(sizeMsg)
		cmds = append(cmds, cmd)

	case ViewDetail:
		var cmd tea.Cmd
		a.detail, cmd = a.detail.Update(msg)
		a.detail, _ = a.detail.Update(sizeMsg)
		cmds = append(cmds, cmd)

	case ViewMcpConfig:
		var cmd tea.Cmd
		a.mcpConfig, cmd = a.mcpConfig.Update(msg)
		a.mcpConfig, _ = a.mcpConfig.Update(sizeMsg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

func (a App) View() string {
	if !a.ready {
		return "Initializing..."
	}

	tabBar := a.tabs.View()

	var body string
	switch a.activeView {
	case ViewDashboard:
		body = a.dashboard.View()
	case ViewMarketplace:
		body = a.marketplace.View()
	case ViewInstalled:
		body = a.installed.View()
	case ViewDetail:
		body = a.detail.View()
	case ViewMcpConfig:
		body = a.mcpConfig.View()
	}

	viewName := "Dashboard"
	var hints []components.KeyHint

	switch a.activeView {
	case ViewDashboard:
		viewName = "Dashboard"
		hints = []components.KeyHint{
			{Key: "tab", Desc: "next tab"},
			{Key: "1/2/3", Desc: "jump to tab"},
			{Key: "q", Desc: "quit"},
		}
	case ViewMarketplace:
		viewName = "Marketplace"
		hints = []components.KeyHint{
			{Key: "↑/↓", Desc: "navigate"},
			{Key: "←/→", Desc: "category"},
			{Key: "space", Desc: "select"},
			{Key: "a", Desc: "select all"},
			{Key: "I", Desc: "batch install"},
			{Key: "/", Desc: "filter"},
			{Key: "enter", Desc: "details"},
			{Key: "tab", Desc: "next tab"},
			{Key: "q", Desc: "quit"},
		}
	case ViewInstalled:
		viewName = "Installed"
		hints = []components.KeyHint{
			{Key: "↑/↓", Desc: "navigate"},
			{Key: "/", Desc: "filter"},
			{Key: "d", Desc: "uninstall"},
			{Key: "u", Desc: "update"},
			{Key: "e", Desc: "edit MCP"},
			{Key: "tab", Desc: "next tab"},
			{Key: "q", Desc: "quit"},
		}
	case ViewDetail:
		viewName = "Detail"
		hints = []components.KeyHint{
			{Key: "i", Desc: "install"},
			{Key: "esc", Desc: "back"},
		}
	case ViewMcpConfig:
		viewName = "MCP Config"
		hints = []components.KeyHint{
			{Key: "tab/↑/↓", Desc: "next field"},
			{Key: "ctrl+s", Desc: "save"},
			{Key: "esc", Desc: "back"},
		}
	}

	a.statusBar = components.StatusBar{
		ViewName: viewName,
		Hints:    hints,
		Version:  a.version,
		Width:    a.width,
	}

	toastLine := a.toast.View(a.width)

	result := tabBar + "\n" + body

	if a.batchPhase != batchNone {
		result += "\n" + a.viewBatchOverlay()
	}

	if toastLine != "" {
		result += "\n" + toastLine
	}
	result += "\n" + a.statusBar.View()

	return result
}

func (a *App) setView(v int) tea.Cmd {
	if v < len(tabNames) {
		a.activeView = v
		a.tabs.SetActive(v)
	}
	if v == ViewMarketplace && !a.marketplaceReady {
		a.marketplaceReady = true
		return a.marketplace.Init()
	}
	return nil
}

func (a *App) initBatchPicker() {
	hasMcp := false
	hasPlugin := false
	for _, item := range a.batchItems {
		if item.Category == "mcps" {
			hasMcp = true
		} else {
			hasPlugin = true
		}
	}
	if hasMcp && hasPlugin {
		a.batchTargets = []string{"cursor", "claude", "codex", "vscode"}
	} else if hasMcp {
		a.batchTargets = []string{"cursor", "claude", "vscode"}
	} else {
		a.batchTargets = []string{"cursor", "claude", "codex"}
	}
	a.batchTargetChecked = make([]bool, len(a.batchTargets))
	for i := range a.batchTargetChecked {
		a.batchTargetChecked[i] = true
	}
	a.batchTargetCursor = 0
	a.batchScopeLocal = true
	a.batchPhase = batchPickTarget
}

func (a App) handleBatchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.batchPhase {
	case batchPickTarget:
		switch msg.String() {
		case "j", "down":
			a.batchTargetCursor = (a.batchTargetCursor + 1) % len(a.batchTargets)
		case "k", "up":
			a.batchTargetCursor = (a.batchTargetCursor - 1 + len(a.batchTargets)) % len(a.batchTargets)
		case " ":
			a.batchTargetChecked[a.batchTargetCursor] = !a.batchTargetChecked[a.batchTargetCursor]
		case "enter":
			a.batchPhase = batchPickScope
		case "esc":
			a.batchPhase = batchNone
			a.batchItems = nil
		}
	case batchPickScope:
		switch msg.String() {
		case "j", "down", "k", "up", "tab":
			a.batchScopeLocal = !a.batchScopeLocal
		case "enter":
			return a.startBatchInstall()
		case "esc":
			a.batchPhase = batchPickTarget
		}
	}
	return a, nil
}

func (a App) startBatchInstall() (tea.Model, tea.Cmd) {
	a.batchPhase = batchInstalling

	var selectedTargets []string
	for i, t := range a.batchTargets {
		if a.batchTargetChecked[i] {
			selectedTargets = append(selectedTargets, t)
		}
	}

	scope := installer.ScopeLocal
	if !a.batchScopeLocal {
		scope = installer.ScopeGlobal
	}

	items := make([]registry.MarketplaceItem, len(a.batchItems))
	copy(items, a.batchItems)

	return a, tea.Batch(a.batchSpinner.Tick, func() tea.Msg {
		return doBatchInstall(items, selectedTargets, scope)
	})
}

func doBatchInstall(items []registry.MarketplaceItem, targets []string, scope installer.Scope) common.BatchInstallCompleteMsg {
	type result struct {
		name string
		err  error
	}
	ch := make(chan result, len(items))
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		go func(it registry.MarketplaceItem) {
			defer wg.Done()
			isMcp := it.Category == "mcps"
			r := views.DoInstall(it, targets, scope, isMcp)
			ch <- result{name: it.Name, err: r.Err}
		}(item)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var succeeded []string
	failed := make(map[string]error)

	for r := range ch {
		if r.err != nil {
			failed[r.name] = r.err
		} else {
			succeeded = append(succeeded, r.name)
		}
	}

	return common.BatchInstallCompleteMsg{Succeeded: succeeded, Failed: failed}
}

func (a App) viewBatchOverlay() string {
	catCounts := make(map[string]int)
	for _, item := range a.batchItems {
		catCounts[item.Category]++
	}

	var sb strings.Builder
	sb.WriteString("\n")

	title := lipgloss.NewStyle().Bold(true).Foreground(common.ColorPrimary).Render(
		fmt.Sprintf("  Batch Install (%d items)", len(a.batchItems)),
	)
	sb.WriteString(title + "\n\n")

	for cat, n := range catCounts {
		icon := common.CategoryIcons[cat]
		badge := common.CategoryBadge(cat)
		sb.WriteString(fmt.Sprintf("  %s %s × %d\n", icon, badge, n))
	}
	sb.WriteString("\n")

	switch a.batchPhase {
	case batchPickTarget:
		sb.WriteString("  " + lipgloss.NewStyle().Bold(true).Render("Select targets:") + "\n")
		for i, t := range a.batchTargets {
			cursor := "  "
			if i == a.batchTargetCursor {
				cursor = lipgloss.NewStyle().Foreground(common.ColorPrimary).Render("▸ ")
			}
			check := "[ ]"
			if a.batchTargetChecked[i] {
				check = common.StyleOK.Render("[✓]")
			}
			sb.WriteString(fmt.Sprintf("  %s%s %s\n", cursor, check, t))
		}
		sb.WriteString(common.StyleDim.Render("\n  space toggle  enter confirm  esc cancel") + "\n")

	case batchPickScope:
		sb.WriteString("  " + lipgloss.NewStyle().Bold(true).Render("Select scope:") + "\n")
		localCursor, globalCursor := "  ", "  "
		if a.batchScopeLocal {
			localCursor = lipgloss.NewStyle().Foreground(common.ColorPrimary).Render("▸ ")
		} else {
			globalCursor = lipgloss.NewStyle().Foreground(common.ColorPrimary).Render("▸ ")
		}
		localRadio, globalRadio := "( )", "( )"
		if a.batchScopeLocal {
			localRadio = common.StyleOK.Render("(●)")
		} else {
			globalRadio = common.StyleOK.Render("(●)")
		}
		sb.WriteString(fmt.Sprintf("  %s%s local (project)\n", localCursor, localRadio))
		sb.WriteString(fmt.Sprintf("  %s%s global (user-wide)\n", globalCursor, globalRadio))
		sb.WriteString(common.StyleDim.Render("\n  ↑/↓ select  enter confirm  esc back") + "\n")

	case batchInstalling:
		sb.WriteString("  " + a.batchSpinner.View() + " Installing " +
			fmt.Sprintf("%d items...", len(a.batchItems)) + "\n")
	}

	overlayStyle := common.StyleOverlay.Width(max(a.width/2, 40))
	return overlayStyle.Render(sb.String())
}
