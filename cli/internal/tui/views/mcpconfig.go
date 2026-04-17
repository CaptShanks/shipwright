package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/CaptShanks/shipwright/cli/internal/installer"
	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
)

type McpConfig struct {
	name    string
	target  string
	scope   installer.Scope
	command string
	args    []string
	envKeys []string
	inputs  []textinput.Model
	cursor  int
	loaded  bool
	saved   bool
	err     error
	width   int
	height  int
}

func NewMcpConfig() McpConfig {
	return McpConfig{scope: installer.ScopeLocal}
}

func (m *McpConfig) Load(name, target, scope string) tea.Cmd {
	m.name = name
	m.target = target
	m.scope = installer.Scope(scope)
	m.loaded = false
	m.saved = false
	m.err = nil

	return func() tea.Msg {
		installers, err := installer.McpForTarget(target)
		if err != nil {
			return common.McpConfigLoadedMsg{Name: name, Err: err}
		}

		for _, inst := range installers {
			cfg, err := inst.ReadConfig(strings.TrimPrefix(name, "mcp:"), installer.Scope(scope))
			if err != nil {
				continue
			}
			return common.McpConfigLoadedMsg{
				Name:    name,
				Target:  target,
				Command: cfg.Command,
				Args:    cfg.Args,
				Env:     cfg.Env,
			}
		}
		return common.McpConfigLoadedMsg{Name: name, Err: fmt.Errorf("config not found for %s in %s", name, target)}
	}
}

func (m McpConfig) Init() tea.Cmd { return nil }

func (m McpConfig) Update(msg tea.Msg) (McpConfig, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case common.McpConfigLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loaded = true
			return m, nil
		}
		m.command = msg.Command
		m.args = msg.Args
		m.loaded = true

		m.envKeys = make([]string, 0, len(msg.Env))
		for k := range msg.Env {
			m.envKeys = append(m.envKeys, k)
		}
		sort.Strings(m.envKeys)

		m.inputs = make([]textinput.Model, len(m.envKeys))
		for i, k := range m.envKeys {
			ti := textinput.New()
			ti.Placeholder = "(empty)"
			ti.SetValue(msg.Env[k])
			ti.CharLimit = 256
			ti.Width = 40
			m.inputs[i] = ti
		}
		m.cursor = 0
		if len(m.inputs) > 0 {
			m.inputs[0].Focus()
			return m, textinput.Blink
		}

	case common.McpConfigSavedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.saved = true
		}

	case tea.KeyMsg:
		if !m.loaded || m.err != nil {
			return m, nil
		}

		switch msg.String() {
		case "tab", "down":
			if len(m.inputs) > 0 {
				m.inputs[m.cursor].Blur()
				m.cursor = (m.cursor + 1) % len(m.inputs)
				m.inputs[m.cursor].Focus()
				return m, textinput.Blink
			}
		case "shift+tab", "up":
			if len(m.inputs) > 0 {
				m.inputs[m.cursor].Blur()
				m.cursor = (m.cursor - 1 + len(m.inputs)) % len(m.inputs)
				m.inputs[m.cursor].Focus()
				return m, textinput.Blink
			}
		case "ctrl+s":
			return m, m.save()
		}
	}

	if m.loaded && len(m.inputs) > 0 {
		var cmd tea.Cmd
		m.inputs[m.cursor], cmd = m.inputs[m.cursor].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m McpConfig) View() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(common.StyleTitle.Render("  MCP Config: "+strings.TrimPrefix(m.name, "mcp:")) + "\n")
	sb.WriteString(common.StyleDim.Render("  Target: "+m.target+"  Scope: "+string(m.scope)) + "\n\n")

	if !m.loaded {
		sb.WriteString(common.StyleDim.Render("  Loading..."))
		return sb.String()
	}

	if m.err != nil {
		sb.WriteString("  " + common.StyleError.Render("Error: "+m.err.Error()) + "\n")
		sb.WriteString(common.StyleDim.Render("\n  [esc] back") + "\n")
		return sb.String()
	}

	sb.WriteString("  " + common.StyleBold.Render("Command:") + " " + m.command + "\n")
	sb.WriteString("  " + common.StyleBold.Render("Args:") + " " + strings.Join(m.args, " ") + "\n\n")

	if len(m.inputs) == 0 {
		sb.WriteString(common.StyleDim.Render("  No environment variables configured.") + "\n")
	} else {
		sb.WriteString("  " + common.StyleBold.Render("Environment Variables:") + "\n\n")
		for i, k := range m.envKeys {
			cursor := "  "
			if i == m.cursor {
				cursor = common.StyleOK.Render("▸ ")
			}
			sb.WriteString(fmt.Sprintf("  %s%s = %s\n", cursor, common.StyleBold.Render(k), m.inputs[i].View()))
		}
	}

	sb.WriteString("\n")

	if m.saved {
		sb.WriteString("  " + common.StyleOK.Render("✓ Saved successfully") + "\n")
	}

	sb.WriteString(common.StyleDim.Render("  [ctrl+s] save  [tab] next field  [esc] back") + "\n")
	return sb.String()
}

func (m McpConfig) save() tea.Cmd {
	name := strings.TrimPrefix(m.name, "mcp:")
	target := m.target
	scope := m.scope
	command := m.command
	args := m.args
	envKeys := m.envKeys
	inputs := m.inputs

	return func() tea.Msg {
		env := make(map[string]string)
		for i, k := range envKeys {
			env[k] = inputs[i].Value()
		}

		cfg := installer.McpConfig{
			Command: command,
			Args:    args,
			Env:     env,
		}

		mcpInstallers, err := installer.McpForTarget(target)
		if err != nil {
			return common.McpConfigSavedMsg{Name: name, Err: err}
		}

		for _, inst := range mcpInstallers {
			if err := inst.SaveConfig(name, cfg, scope); err != nil {
				return common.McpConfigSavedMsg{Name: name, Err: err}
			}
		}
		return common.McpConfigSavedMsg{Name: name}
	}
}
