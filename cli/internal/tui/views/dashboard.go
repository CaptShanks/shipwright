package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CaptShanks/shipwright/cli/internal/state"
	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
)

type Dashboard struct {
	installations []state.Installation
	width, height int
	loaded        bool
}

func NewDashboard() Dashboard {
	return Dashboard{}
}

func (d Dashboard) Init() tea.Cmd {
	return d.loadState
}

func (d Dashboard) Update(msg tea.Msg) (Dashboard, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
	case common.StateFetchedMsg:
		d.loaded = true
		if msg.Err == nil {
			d.installations = msg.Installations
		}
	}
	return d, nil
}

func (d Dashboard) View() string {
	if !d.loaded {
		return common.StyleDim.Render("  Loading...")
	}

	counts := map[string]struct{ agents, skills, mcps int }{
		"cursor": {},
		"claude": {},
		"codex":  {},
	}

	for _, inst := range d.installations {
		c := counts[inst.Target]
		if strings.HasPrefix(inst.Plugin, "mcp:") {
			c.mcps++
		} else if strings.Contains(inst.Plugin, "skill") {
			c.skills++
		} else {
			c.agents++
		}
		counts[inst.Target] = c
	}

	var cards []string
	for _, tool := range []string{"cursor", "claude", "codex"} {
		c := counts[tool]
		total := c.agents + c.skills + c.mcps

		title := common.StyleBold.Render(strings.ToUpper(tool[:1]) + tool[1:])
		body := fmt.Sprintf(
			"%s agents  %s skills  %s MCPs",
			common.StyleOK.Render(fmt.Sprintf("%d", c.agents)),
			common.StyleAccent.Render(fmt.Sprintf("%d", c.skills)),
			lipgloss.NewStyle().Foreground(common.ColorPrimary).Render(fmt.Sprintf("%d", c.mcps)),
		)

		style := common.StyleCard
		if total > 0 {
			style = common.StyleCardActive
		}

		cardWidth := max((d.width-10)/3, 20)
		card := style.Width(cardWidth).Render(title + "\n" + body)
		cards = append(cards, card)
	}

	cardRow := lipgloss.JoinHorizontal(lipgloss.Top, cards...)

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(common.StyleTitle.Render("  Dashboard") + "\n\n")
	sb.WriteString("  " + cardRow + "\n\n")

	recent := d.recentActivity(5)
	if len(recent) > 0 {
		sb.WriteString(common.StyleTitle.Render("  Recent Activity") + "\n\n")
		for _, inst := range recent {
			name := inst.Plugin
			if strings.HasPrefix(name, "mcp:") {
				name = common.StyleBadge.Render("MCP") + " " + strings.TrimPrefix(name, "mcp:")
			}
			line := fmt.Sprintf("  %s  %s  %s  %s",
				name,
				common.StyleDim.Render(inst.Target),
				common.StyleDim.Render(inst.Scope),
				common.StyleDim.Render(inst.InstalledAt.Format("Jan 02 15:04")),
			)
			sb.WriteString(line + "\n")
		}
	} else {
		sb.WriteString(common.StyleDim.Render("  No plugins installed yet. Press 2 to browse the marketplace.") + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(common.StyleDim.Render("  [2] marketplace  [3] installed  [?] help") + "\n")

	return sb.String()
}

func (d Dashboard) recentActivity(n int) []state.Installation {
	sorted := make([]state.Installation, len(d.installations))
	copy(sorted, d.installations)

	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j].InstalledAt.After(sorted[j-1].InstalledAt); j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}

	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}

func (d Dashboard) loadState() tea.Msg {
	store := state.NewStore()
	installs, err := store.List("", "", "")
	return common.StateFetchedMsg{Installations: installs, Err: err}
}
