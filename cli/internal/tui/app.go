package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

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

type App struct {
	tabs        components.TabBar
	statusBar   components.StatusBar
	toast       components.Toast
	dashboard   views.Dashboard
	marketplace views.Marketplace
	installed   views.Installed
	detail      views.Detail
	mcpConfig   views.McpConfig
	activeView  int
	prevView    int
	width       int
	height      int
	version     string
	ready       bool
}

func NewApp(version string) App {
	return App{
		tabs:        components.NewTabBar(tabNames),
		toast:       components.NewToast(),
		dashboard:   views.NewDashboard(),
		marketplace: views.NewMarketplace(),
		installed:   views.NewInstalled(),
		detail:      views.NewDetail(),
		mcpConfig:   views.NewMcpConfig(),
		version:     version,
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.dashboard.Init(),
		a.marketplace.Init(),
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
				a.activeView = a.tabs.Active
			case key.Matches(msg, Keys.ShiftTab):
				a.tabs.Prev()
				a.activeView = a.tabs.Active
			case key.Matches(msg, Keys.Num1):
				a.setView(ViewDashboard)
			case key.Matches(msg, Keys.Num2):
				a.setView(ViewMarketplace)
			case key.Matches(msg, Keys.Num3):
				a.setView(ViewInstalled)
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

		if a.marketplace.SelectedItem != nil {
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
	hints := " tab navigate  ? help  q quit"
	if a.activeView < len(tabNames) {
		viewName = tabNames[a.activeView]
	} else if a.activeView == ViewDetail {
		viewName = "Detail"
		hints = " esc back  i install  q quit"
	} else if a.activeView == ViewMcpConfig {
		viewName = "MCP Config"
		hints = " esc back  ctrl+s save  tab next"
	}

	a.statusBar = components.StatusBar{
		ViewName: viewName,
		HintKeys: hints,
		Version:  a.version,
		Width:    a.width,
	}

	toastLine := a.toast.View(a.width)

	result := tabBar + "\n" + body
	if toastLine != "" {
		result += "\n" + toastLine
	}
	result += "\n" + a.statusBar.View()

	return result
}

func (a *App) setView(v int) {
	if v < len(tabNames) {
		a.activeView = v
		a.tabs.SetActive(v)
	}
}
