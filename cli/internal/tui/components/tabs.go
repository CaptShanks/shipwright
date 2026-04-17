package components

import (
	"strings"

	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
	"github.com/charmbracelet/lipgloss"
)

type TabBar struct {
	Tabs   []string
	Active int
	Width  int
}

func NewTabBar(tabs []string) TabBar {
	return TabBar{Tabs: tabs}
}

func (t *TabBar) SetActive(idx int) {
	if idx >= 0 && idx < len(t.Tabs) {
		t.Active = idx
	}
}

func (t *TabBar) Next() {
	t.Active = (t.Active + 1) % len(t.Tabs)
}

func (t *TabBar) Prev() {
	t.Active = (t.Active - 1 + len(t.Tabs)) % len(t.Tabs)
}

func (t TabBar) View() string {
	var tabs []string
	for i, name := range t.Tabs {
		if i == t.Active {
			tabs = append(tabs, common.StyleActiveTab.Render(name))
		} else {
			tabs = append(tabs, common.StyleInactiveTab.Render(name))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	border := common.StyleDim.Render(strings.Repeat("─", max(t.Width, lipgloss.Width(row))))

	return row + "\n" + border
}
