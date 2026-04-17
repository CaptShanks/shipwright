package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CaptShanks/shipwright/cli/internal/installer"
	"github.com/CaptShanks/shipwright/cli/internal/state"
	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
	"github.com/CaptShanks/shipwright/cli/internal/tui/components"
)

type Installed struct {
	filter        textinput.Model
	confirm       components.ConfirmDialog
	installations []state.Installation
	filtered      []state.Installation
	filtering     bool
	loaded        bool
	cursor        int
	scrollOffset  int
	width, height int
	pendingAction string
	pendingPlugin string
	pendingTarget string
	pendingScope  string
	pendingPath   string
}

func NewInstalled() Installed {
	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.CharLimit = 64

	return Installed{filter: ti}
}

func (v Installed) Init() tea.Cmd {
	return v.loadState
}

func (v Installed) Update(msg tea.Msg) (Installed, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case common.StateFetchedMsg:
		v.loaded = true
		if msg.Err == nil {
			v.installations = msg.Installations
			v.applyFilter()
		}

	case common.UninstallCompleteMsg:
		return v, v.loadState

	case common.UpdateCompleteMsg:
		return v, v.loadState

	case tea.KeyMsg:
		if v.confirm.Visible {
			return v.handleConfirmKey(msg)
		}

		if v.filtering {
			switch msg.String() {
			case "esc":
				v.filtering = false
				v.filter.Blur()
				v.filter.SetValue("")
				v.applyFilter()
				return v, nil
			case "enter":
				v.filtering = false
				v.filter.Blur()
				return v, nil
			default:
				var cmd tea.Cmd
				v.filter, cmd = v.filter.Update(msg)
				v.applyFilter()
				return v, cmd
			}
		}

		switch msg.String() {
		case "/":
			v.filtering = true
			v.filter.Focus()
			return v, textinput.Blink
		case "j", "down":
			if v.cursor < len(v.filtered)-1 {
				v.cursor++
				v.ensureVisible()
			}
		case "k", "up":
			if v.cursor > 0 {
				v.cursor--
				v.ensureVisible()
			}
		case "d", "delete":
			if len(v.filtered) > 0 {
				inst := v.filtered[v.cursor]
				v.pendingAction = "uninstall"
				v.pendingPlugin = inst.Plugin
				v.pendingTarget = inst.Target
				v.pendingScope = inst.Scope
				if inst.Scope == "local" {
					v.pendingPath = installer.ProjectRoot()
				}
				v.confirm = components.ConfirmDialog{
					Title:   "Uninstall " + inst.Plugin + "?",
					Message: fmt.Sprintf("Remove %s from %s (%s)?", inst.Plugin, inst.Target, inst.Scope),
					Visible: true,
				}
			}
		case "u":
			if len(v.filtered) > 0 {
				inst := v.filtered[v.cursor]
				return v, v.doUpdate(inst.Plugin, inst.Target, inst.Scope)
			}
		case "e":
			if len(v.filtered) > 0 {
				inst := v.filtered[v.cursor]
				if strings.HasPrefix(inst.Plugin, "mcp:") {
					row := []string{inst.Plugin, inst.Target, inst.Scope}
					return v, func() tea.Msg {
						return common.NavigateMsg{View: 4, Context: row}
					}
				}
			}
		}
	}

	return v, nil
}

func (v Installed) handleConfirmKey(msg tea.KeyMsg) (Installed, tea.Cmd) {
	switch msg.String() {
	case "y":
		v.confirm.Visible = false
		if v.pendingAction == "uninstall" {
			return v, v.doUninstall(v.pendingPlugin, v.pendingTarget, v.pendingScope, v.pendingPath)
		}
	case "n", "esc":
		v.confirm.Visible = false
	}
	return v, nil
}

func (v Installed) View() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(common.StyleTitle.Render("  Installed Plugins") + "\n\n")

	if !v.loaded {
		sb.WriteString(common.StyleDim.Render("  Loading..."))
		return sb.String()
	}

	if v.filtering {
		sb.WriteString("  " + v.filter.View() + "\n\n")
	}

	if len(v.filtered) == 0 {
		sb.WriteString(common.StyleDim.Render("  No plugins installed. Press 2 to browse the marketplace.") + "\n")
		return sb.String()
	}

	// Summary counts
	agents, skills, mcps := 0, 0, 0
	for _, inst := range v.filtered {
		switch {
		case strings.HasPrefix(inst.Plugin, "mcp:"):
			mcps++
		case strings.Contains(inst.Plugin, "skill"):
			skills++
		default:
			agents++
		}
	}
	summary := fmt.Sprintf("  %s  %s  %s  %s",
		common.StyleDim.Render(fmt.Sprintf("%d total", len(v.filtered))),
		lipgloss.NewStyle().Foreground(common.ColorAgent).Render(fmt.Sprintf("%d agents", agents)),
		lipgloss.NewStyle().Foreground(common.ColorSkill).Render(fmt.Sprintf("%d skills", skills)),
		lipgloss.NewStyle().Foreground(common.ColorMCP).Render(fmt.Sprintf("%d MCPs", mcps)),
	)
	sb.WriteString(summary + "\n\n")

	visibleRows := v.height - 12
	if visibleRows < 3 {
		visibleRows = 3
	}

	end := v.scrollOffset + visibleRows
	if end > len(v.filtered) {
		end = len(v.filtered)
	}

	for i := v.scrollOffset; i < end; i++ {
		inst := v.filtered[i]
		selected := i == v.cursor
		sb.WriteString(v.renderRow(inst, selected) + "\n")
	}

	if v.confirm.Visible {
		return v.confirm.View(v.width, v.height)
	}

	return sb.String()
}

func (v Installed) renderRow(inst state.Installation, selected bool) string {
	cat := "agents"
	name := inst.Plugin
	if strings.HasPrefix(inst.Plugin, "mcp:") {
		cat = "mcps"
		name = strings.TrimPrefix(inst.Plugin, "mcp:")
	} else if strings.Contains(inst.Plugin, "skill") {
		cat = "skills"
	}

	catColor := common.CategoryColor(cat)
	badge := common.CategoryBadge(cat)
	icon := common.CategoryIcons[cat]

	cursor := "  "
	nameStyle := lipgloss.NewStyle().Bold(true)
	rowBg := lipgloss.NewStyle()
	if selected {
		cursor = lipgloss.NewStyle().Foreground(catColor).Render("▸ ")
		nameStyle = nameStyle.Foreground(catColor)
		rowBg = rowBg.Background(lipgloss.Color("236"))
	}

	targetBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("240")).
		Padding(0, 1).
		Render(inst.Target)

	scopeStr := common.StyleDim.Render(inst.Scope)
	verStr := common.StyleDim.Render("v" + inst.Version)
	filesStr := common.StyleDim.Render(fmt.Sprintf("%d files", len(inst.Files)))
	timeStr := common.StyleDim.Render(inst.InstalledAt.Format("Jan 02 15:04"))

	line1 := fmt.Sprintf(" %s%s %s  %s  %s", cursor, icon, nameStyle.Render(name), badge, targetBadge)
	line2 := fmt.Sprintf("       %s  %s  %s  %s", verStr, scopeStr, filesStr, timeStr)

	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(
		"     " + strings.Repeat("─", max(v.width-10, 20)),
	)

	if selected {
		return rowBg.Width(v.width - 2).Render(line1+"\n"+line2) + "\n" + separator
	}
	return line1 + "\n" + line2 + "\n" + separator
}

func (v *Installed) ensureVisible() {
	visibleRows := v.height - 12
	if visibleRows < 3 {
		visibleRows = 3
	}
	if v.cursor < v.scrollOffset {
		v.scrollOffset = v.cursor
	} else if v.cursor >= v.scrollOffset+visibleRows {
		v.scrollOffset = v.cursor - visibleRows + 1
	}
}

func (v *Installed) applyFilter() {
	query := strings.ToLower(v.filter.Value())
	v.filtered = nil

	for _, inst := range v.installations {
		if query != "" {
			combined := strings.ToLower(inst.Plugin + " " + inst.Target + " " + inst.Scope)
			if !strings.Contains(combined, query) {
				continue
			}
		}
		v.filtered = append(v.filtered, inst)
	}

	if v.cursor >= len(v.filtered) {
		v.cursor = max(len(v.filtered)-1, 0)
	}
}

func (v Installed) loadState() tea.Msg {
	store := state.NewStore()
	installs, err := store.List("", "", "")
	return common.StateFetchedMsg{Installations: installs, Err: err}
}

func (v Installed) doUninstall(plugin, target, scope, projectPath string) tea.Cmd {
	return func() tea.Msg {
		store := state.NewStore()

		if strings.HasPrefix(plugin, "mcp:") {
			mcpName := strings.TrimPrefix(plugin, "mcp:")
			mcpInstallers, _ := installer.McpForTarget(target)
			sc := installer.Scope(scope)
			for _, inst := range mcpInstallers {
				_ = inst.Remove(mcpName, sc)
			}
			store.Remove(plugin, target, scope, projectPath)
			return common.UninstallCompleteMsg{Plugin: plugin, Target: target}
		}

		targetInstallers, _ := installer.ForTarget(target)
		for _, inst := range targetInstallers {
			files, _ := store.Remove(plugin, inst.Name(), scope, projectPath)
			_ = inst.UninstallFiles(files)
		}

		return common.UninstallCompleteMsg{Plugin: plugin, Target: target}
	}
}

func (v Installed) doUpdate(plugin, target, scope string) tea.Cmd {
	return func() tea.Msg {
		return common.UpdateCompleteMsg{Plugin: plugin, Target: target}
	}
}
