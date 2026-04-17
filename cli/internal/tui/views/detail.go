package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CaptShanks/shipwright/cli/internal/installer"
	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/state"
	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
)

type installPhase int

const (
	phaseNone installPhase = iota
	phasePickTarget
	phasePickScope
	phaseInstalling
)

type Detail struct {
	item          *registry.MarketplaceItem
	manifest      *registry.PluginManifest
	mcpManifest   *registry.McpManifest
	spinner       spinner.Model
	loading       bool
	installing    bool
	phase         installPhase
	err           error
	width         int
	height        int
	installations []state.Installation

	targets       []string
	targetChecked []bool
	targetCursor  int
	scopeLocal    bool
}

func NewDetail() Detail {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return Detail{spinner: sp, scopeLocal: true}
}

func (d *Detail) SetItem(item *registry.MarketplaceItem) tea.Cmd {
	d.item = item
	d.manifest = nil
	d.mcpManifest = nil
	d.loading = true
	d.err = nil
	d.phase = phaseNone
	d.installing = false

	source := item.Source

	if item.Category == "mcps" {
		d.targets = []string{"cursor", "claude", "vscode"}
	} else {
		d.targets = []string{"cursor", "claude", "codex"}
	}
	d.targetChecked = make([]bool, len(d.targets))
	for i := range d.targetChecked {
		d.targetChecked[i] = true
	}
	d.targetCursor = 0

	return tea.Batch(d.spinner.Tick, d.fetchManifest(source, item.Category))
}

func (d Detail) Init() tea.Cmd {
	return nil
}

func (d Detail) Update(msg tea.Msg) (Detail, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

	case common.ManifestFetchedMsg:
		d.loading = false
		if msg.Err != nil {
			d.err = msg.Err
		} else {
			d.manifest = msg.Manifest
		}

	case common.McpManifestFetchedMsg:
		d.loading = false
		if msg.Err != nil {
			d.err = msg.Err
		} else {
			d.mcpManifest = msg.Manifest
		}

	case common.StateFetchedMsg:
		if msg.Err == nil {
			d.installations = msg.Installations
		}

	case common.InstallCompleteMsg:
		d.installing = false
		if msg.Err != nil {
			d.err = msg.Err
		}
		d.phase = phaseNone
		return d, d.refreshState

	case spinner.TickMsg:
		if d.loading || d.installing {
			var cmd tea.Cmd
			d.spinner, cmd = d.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		switch d.phase {
		case phasePickTarget:
			return d.updateTargetPicker(msg)
		case phasePickScope:
			return d.updateScopePicker(msg)
		default:
			switch msg.String() {
			case "i":
				if d.item != nil && !d.loading {
					d.phase = phasePickTarget
					d.targetCursor = 0
				}
			}
		}
	}

	return d, tea.Batch(cmds...)
}

func (d Detail) updateTargetPicker(msg tea.KeyMsg) (Detail, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		d.targetCursor = (d.targetCursor + 1) % len(d.targets)
	case "k", "up":
		d.targetCursor = (d.targetCursor - 1 + len(d.targets)) % len(d.targets)
	case " ":
		d.targetChecked[d.targetCursor] = !d.targetChecked[d.targetCursor]
	case "enter":
		d.phase = phasePickScope
	case "esc":
		d.phase = phaseNone
	}
	return d, nil
}

func (d Detail) updateScopePicker(msg tea.KeyMsg) (Detail, tea.Cmd) {
	switch msg.String() {
	case "j", "down", "k", "up", "tab":
		d.scopeLocal = !d.scopeLocal
	case "enter":
		return d.startInstall()
	case "esc":
		d.phase = phasePickTarget
	}
	return d, nil
}

func (d Detail) startInstall() (Detail, tea.Cmd) {
	d.phase = phaseInstalling
	d.installing = true

	var selectedTargets []string
	for i, t := range d.targets {
		if d.targetChecked[i] {
			selectedTargets = append(selectedTargets, t)
		}
	}

	scope := installer.ScopeLocal
	if !d.scopeLocal {
		scope = installer.ScopeGlobal
	}

	item := *d.item
	isMcp := item.Category == "mcps"

	return d, tea.Batch(d.spinner.Tick, func() tea.Msg {
		return doInstall(item, selectedTargets, scope, isMcp)
	})
}

func (d Detail) View() string {
	if d.item == nil {
		return common.StyleDim.Render("\n  No item selected")
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(common.StyleTitle.Render("  "+d.item.Name) + "  ")
	sb.WriteString(common.StyleDim.Render("v"+d.item.Version) + "  ")
	sb.WriteString(common.StyleBadge.Render(d.item.Category) + "\n\n")
	sb.WriteString("  " + d.item.Description + "\n\n")

	if d.loading {
		sb.WriteString("  " + d.spinner.View() + " Loading details...\n")
		return sb.String()
	}

	if d.err != nil {
		sb.WriteString("  " + common.StyleError.Render("Error: "+d.err.Error()) + "\n")
	}

	if d.isInstalled() {
		sb.WriteString("  " + common.StyleBadge.Render("INSTALLED") + "\n\n")
	}

	if d.manifest != nil {
		if len(d.manifest.Agents) > 0 {
			sb.WriteString("  " + common.StyleBold.Render("Agents:") + " " + strings.Join(d.manifest.Agents, ", ") + "\n")
		}
		if len(d.manifest.Skills) > 0 {
			sb.WriteString("  " + common.StyleBold.Render("Skills:") + " " + strings.Join(d.manifest.Skills, ", ") + "\n")
		}
	}

	if d.mcpManifest != nil {
		sb.WriteString("  " + common.StyleBold.Render("Command:") + " " + d.mcpManifest.Command + " " + strings.Join(d.mcpManifest.Args, " ") + "\n")
		if len(d.mcpManifest.Env) > 0 {
			sb.WriteString("  " + common.StyleBold.Render("Env vars:") + "\n")
			for k, v := range d.mcpManifest.Env {
				val := v
				if val == "" {
					val = common.StyleWarn.Render("(required)")
				}
				sb.WriteString(fmt.Sprintf("    %s = %s\n", k, val))
			}
		}
	}

	if len(d.item.Tags) > 0 {
		sb.WriteString("\n  " + common.StyleDim.Render("Tags: "+strings.Join(d.item.Tags, ", ")) + "\n")
	}

	sb.WriteString("\n")

	switch d.phase {
	case phasePickTarget:
		sb.WriteString(d.viewTargetPicker())
	case phasePickScope:
		sb.WriteString(d.viewScopePicker())
	case phaseInstalling:
		sb.WriteString("  " + d.spinner.View() + " Installing...\n")
	default:
		sb.WriteString(common.StyleDim.Render("  [i] install  [esc] back") + "\n")
	}

	return sb.String()
}

func (d Detail) viewTargetPicker() string {
	var sb strings.Builder
	sb.WriteString("  " + common.StyleBold.Render("Select targets:") + "\n")
	for i, t := range d.targets {
		cursor := "  "
		if i == d.targetCursor {
			cursor = lipgloss.NewStyle().Foreground(common.ColorPrimary).Render("▸ ")
		}
		check := "[ ]"
		if d.targetChecked[i] {
			check = common.StyleOK.Render("[✓]")
		}
		sb.WriteString(fmt.Sprintf("  %s%s %s\n", cursor, check, t))
	}
	sb.WriteString(common.StyleDim.Render("\n  space toggle  enter confirm  esc cancel") + "\n")
	return sb.String()
}

func (d Detail) viewScopePicker() string {
	var sb strings.Builder
	sb.WriteString("  " + common.StyleBold.Render("Select scope:") + "\n")

	localCursor := "  "
	globalCursor := "  "
	if d.scopeLocal {
		localCursor = lipgloss.NewStyle().Foreground(common.ColorPrimary).Render("▸ ")
	} else {
		globalCursor = lipgloss.NewStyle().Foreground(common.ColorPrimary).Render("▸ ")
	}

	localRadio := "( )"
	globalRadio := "( )"
	if d.scopeLocal {
		localRadio = common.StyleOK.Render("(●)")
	} else {
		globalRadio = common.StyleOK.Render("(●)")
	}

	sb.WriteString(fmt.Sprintf("  %s%s local (project)\n", localCursor, localRadio))
	sb.WriteString(fmt.Sprintf("  %s%s global (user-wide)\n", globalCursor, globalRadio))
	sb.WriteString(common.StyleDim.Render("\n  ↑/↓ select  enter confirm  esc back") + "\n")
	return sb.String()
}

func (d Detail) isInstalled() bool {
	if d.item == nil {
		return false
	}
	name := d.item.Name
	if d.item.Category == "mcps" {
		name = "mcp:" + name
	}
	for _, inst := range d.installations {
		if inst.Plugin == name {
			return true
		}
	}
	return false
}

func (d Detail) fetchManifest(source, category string) func() tea.Msg {
	return func() tea.Msg {
		client := registry.NewClient()
		if category == "mcps" {
			m, err := client.FetchMcpManifest(source)
			return common.McpManifestFetchedMsg{Manifest: m, Err: err}
		}
		m, err := client.FetchPluginManifest(source)
		return common.ManifestFetchedMsg{Manifest: m, Err: err}
	}
}

func (d Detail) refreshState() tea.Msg {
	store := state.NewStore()
	installs, err := store.List("", "", "")
	return common.StateFetchedMsg{Installations: installs, Err: err}
}

func doInstall(item registry.MarketplaceItem, targets []string, scope installer.Scope, isMcp bool) common.InstallCompleteMsg {
	client := registry.NewClient()
	store := state.NewStore()

	if isMcp {
		manifest, err := client.FetchMcpManifest(item.Source)
		if err != nil {
			return common.InstallCompleteMsg{Plugin: item.Name, Err: err}
		}

		cfg := installer.McpConfig{
			Command: manifest.Command,
			Args:    manifest.Args,
			Env:     manifest.Env,
		}

		for _, target := range targets {
			mcpInstallers, err := installer.McpForTarget(target)
			if err != nil {
				continue
			}
			for _, inst := range mcpInstallers {
				if err := inst.Install(manifest.Name, cfg, scope, false); err != nil {
					continue
				}
				_ = store.Record(state.Installation{
					Plugin:  "mcp:" + item.Name,
					Version: item.Version,
					Target:  inst.Name(),
					Scope:   string(scope),
					Files:   []string{inst.ConfigPath(scope)},
				})
			}
		}
		return common.InstallCompleteMsg{Plugin: item.Name}
	}

	manifest, err := client.FetchPluginManifest(item.Source)
	if err != nil {
		return common.InstallCompleteMsg{Plugin: item.Name, Err: err}
	}

	agentContents := make(map[string][]byte)
	for _, agentPath := range manifest.Agents {
		agentFile := strings.TrimPrefix(agentPath, "agents/")
		content, err := client.FetchFileContent("_agents/" + agentFile)
		if err != nil {
			return common.InstallCompleteMsg{Plugin: item.Name, Err: err}
		}
		name := strings.TrimSuffix(agentFile, ".md")
		agentContents[name] = content
	}

	skillContents := make(map[string]map[string][]byte)
	for _, skillPath := range manifest.Skills {
		skillName := strings.TrimPrefix(skillPath, "skills/")
		files, err := client.FetchSkillTree("_skills/" + skillName)
		if err != nil {
			return common.InstallCompleteMsg{Plugin: item.Name, Err: err}
		}
		skillContents[skillName] = files
	}

	for _, target := range targets {
		targetInstallers, err := installer.ForTarget(target)
		if err != nil {
			continue
		}
		for _, inst := range targetInstallers {
			var allFiles []string
			for name, content := range agentContents {
				path, err := inst.InstallAgent(name, content, scope)
				if err != nil {
					continue
				}
				allFiles = append(allFiles, path)
			}
			for name, files := range skillContents {
				paths, err := inst.InstallSkill(name, files, scope)
				if err != nil {
					continue
				}
				allFiles = append(allFiles, paths...)
			}
			_ = store.Record(state.Installation{
				Plugin:  item.Name,
				Version: manifest.Version,
				Target:  inst.Name(),
				Scope:   string(scope),
				Files:   allFiles,
			})
		}
	}

	return common.InstallCompleteMsg{Plugin: item.Name}
}
