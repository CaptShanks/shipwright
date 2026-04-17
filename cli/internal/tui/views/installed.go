package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CaptShanks/shipwright/cli/internal/installer"
	"github.com/CaptShanks/shipwright/cli/internal/state"
	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
	"github.com/CaptShanks/shipwright/cli/internal/tui/components"
)

type Installed struct {
	table         table.Model
	filter        textinput.Model
	confirm       components.ConfirmDialog
	installations []state.Installation
	filtered      []state.Installation
	filtering     bool
	loaded        bool
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

	cols := []table.Column{
		{Title: "Plugin", Width: 24},
		{Title: "Target", Width: 10},
		{Title: "Scope", Width: 8},
		{Title: "Version", Width: 10},
		{Title: "Files", Width: 6},
		{Title: "Installed", Width: 16},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(true).
		Foreground(common.ColorPrimary).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(common.ColorDim)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("0")).
		Background(common.ColorPrimary)
	t.SetStyles(s)

	return Installed{table: t, filter: ti}
}

func (v Installed) Init() tea.Cmd {
	return v.loadState
}

func (v Installed) Update(msg tea.Msg) (Installed, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.table.SetWidth(msg.Width - 4)
		v.table.SetHeight(msg.Height - 10)

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
		case "d", "delete":
			if len(v.filtered) > 0 {
				row := v.table.SelectedRow()
				if row != nil {
					v.pendingAction = "uninstall"
					v.pendingPlugin = row[0]
					v.pendingTarget = row[1]
					v.pendingScope = row[2]
					if v.pendingScope == "local" {
						v.pendingPath = installer.ProjectRoot()
					}
					v.confirm = components.ConfirmDialog{
						Title:   "Uninstall " + row[0] + "?",
						Message: fmt.Sprintf("Remove %s from %s (%s)?", row[0], row[1], row[2]),
						Visible: true,
					}
				}
			}
		case "u":
			if len(v.filtered) > 0 {
				row := v.table.SelectedRow()
				if row != nil {
					return v, v.doUpdate(row[0], row[1], row[2])
				}
			}
		case "e":
			if len(v.filtered) > 0 {
				row := v.table.SelectedRow()
				if row != nil && strings.HasPrefix(row[0], "mcp:") {
					return v, func() tea.Msg {
						return common.NavigateMsg{View: 4, Context: row}
					}
				}
			}
		}
	}

	if v.loaded && !v.filtering {
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return v, cmd
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
	} else {
		sb.WriteString("  " + v.table.View() + "\n")
	}

	if v.confirm.Visible {
		return v.confirm.View(v.width, v.height)
	}

	return sb.String()
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

	rows := make([]table.Row, len(v.filtered))
	for i, inst := range v.filtered {
		rows[i] = table.Row{
			inst.Plugin,
			inst.Target,
			inst.Scope,
			inst.Version,
			fmt.Sprintf("%d", len(inst.Files)),
			inst.InstalledAt.Format("Jan 02 15:04"),
		}
	}
	v.table.SetRows(rows)
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
