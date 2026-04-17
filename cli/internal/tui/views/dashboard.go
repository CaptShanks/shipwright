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

	toolIcons := map[string]string{
		"cursor": "⌨",
		"claude": "🧠",
		"codex":  "📟",
	}

	var cards []string
	for _, tool := range []string{"cursor", "claude", "codex"} {
		c := counts[tool]
		total := c.agents + c.skills + c.mcps

		icon := toolIcons[tool]
		title := common.StyleBold.Render(icon + " " + strings.ToUpper(tool[:1]) + tool[1:])

		agentLine := fmt.Sprintf("  %s %s  %s",
			common.CategoryIcons["agents"],
			lipgloss.NewStyle().Foreground(common.ColorAgent).Render(fmt.Sprintf("%d", c.agents)),
			lipgloss.NewStyle().Foreground(common.ColorAgent).Render("Agents"),
		)
		skillLine := fmt.Sprintf("  %s %s  %s",
			common.CategoryIcons["skills"],
			lipgloss.NewStyle().Foreground(common.ColorSkill).Render(fmt.Sprintf("%d", c.skills)),
			lipgloss.NewStyle().Foreground(common.ColorSkill).Render("Skills"),
		)
		mcpLine := fmt.Sprintf("  %s %s  %s",
			common.CategoryIcons["mcps"],
			lipgloss.NewStyle().Foreground(common.ColorMCP).Render(fmt.Sprintf("%d", c.mcps)),
			lipgloss.NewStyle().Foreground(common.ColorMCP).Render("MCPs"),
		)

		style := common.StyleCard
		if total > 0 {
			style = common.StyleCardActive
		}

		cardWidth := max((d.width-10)/3, 20)
		card := style.Width(cardWidth).Render(title + "\n" + agentLine + "\n" + skillLine + "\n" + mcpLine)
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
			cat := "agents"
			name := inst.Plugin
			if strings.HasPrefix(name, "mcp:") {
				cat = "mcps"
				name = strings.TrimPrefix(name, "mcp:")
			} else if strings.Contains(name, "skill") {
				cat = "skills"
			}
			icon := common.CategoryIcons[cat]
			badge := common.CategoryBadge(cat)
			catColor := common.CategoryColor(cat)

			nameStyled := lipgloss.NewStyle().Bold(true).Foreground(catColor).Render(name)

			targetBadge := lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("240")).
				Padding(0, 1).
				Render(inst.Target)

			line := fmt.Sprintf("  %s %s  %s  %s  %s  %s",
				icon,
				nameStyled,
				badge,
				targetBadge,
				common.StyleDim.Render(inst.Scope),
				common.StyleDim.Render(inst.InstalledAt.Format("Jan 02 15:04")),
			)
			sb.WriteString(line + "\n")
		}
	} else {
		sb.WriteString(common.StyleDim.Render("  No plugins installed yet. Press 2 to browse the marketplace.") + "\n")
	}

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
