package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CaptShanks/shipwright/cli/internal/registry"
	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
)

var categories = []string{"all", "agents", "skills", "bundles", "mcps"}

type marketplaceItem struct {
	item registry.MarketplaceItem
}

func (m marketplaceItem) FilterValue() string {
	return m.item.Name + " " + m.item.Description + " " + strings.Join(m.item.Tags, " ")
}

type marketplaceDelegate struct{}

func (d marketplaceDelegate) Height() int                             { return 2 }
func (d marketplaceDelegate) Spacing() int                            { return 0 }
func (d marketplaceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d marketplaceDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	mi, ok := item.(marketplaceItem)
	if !ok {
		return
	}

	name := mi.item.Name
	desc := mi.item.Description
	cat := mi.item.Category
	ver := mi.item.Version

	selected := index == m.Index()

	nameStyle := lipgloss.NewStyle()
	if selected {
		nameStyle = nameStyle.Bold(true).Foreground(common.ColorPrimary)
	}

	catBadge := common.StyleDim.Render("[" + cat + "]")
	verStr := common.StyleDim.Render("v" + ver)

	line1 := fmt.Sprintf("  %s %s %s", nameStyle.Render(name), catBadge, verStr)

	maxDesc := m.Width() - 6
	if maxDesc > 0 && len(desc) > maxDesc {
		desc = desc[:maxDesc-3] + "..."
	}
	line2 := common.StyleDim.Render("    " + desc)

	if selected {
		cursor := lipgloss.NewStyle().Foreground(common.ColorPrimary).Render("▸")
		line1 = cursor + line1[1:]
	}

	fmt.Fprintf(w, "%s\n%s", line1, line2)
}

type Marketplace struct {
	list         list.Model
	spinner      spinner.Model
	allItems     []registry.MarketplaceItem
	loading      bool
	loaded       bool
	catIndex     int
	width        int
	height       int
	err          error
	SelectedItem *registry.MarketplaceItem
}

func NewMarketplace() Marketplace {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	l := list.New(nil, marketplaceDelegate{}, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()
	l.KeyMap.Quit = key.NewBinding(key.WithDisabled())

	return Marketplace{
		list:    l,
		spinner: sp,
		loading: true,
	}
}

func (m Marketplace) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchMarketplace)
}

func (m Marketplace) Update(msg tea.Msg) (Marketplace, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)

	case common.MarketplaceFetchedMsg:
		m.loading = false
		m.loaded = true
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.allItems = msg.Marketplace.Plugins
		m.applyFilter()

	case tea.KeyMsg:
		if !m.list.SettingFilter() {
			switch msg.String() {
			case "left", "h":
				m.catIndex = (m.catIndex - 1 + len(categories)) % len(categories)
				m.applyFilter()
			case "right", "l":
				m.catIndex = (m.catIndex + 1) % len(categories)
				m.applyFilter()
			case "enter":
				if item, ok := m.list.SelectedItem().(marketplaceItem); ok {
					selected := item.item
					m.SelectedItem = &selected
				}
				return m, nil
			}
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if m.loaded {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Marketplace) View() string {
	var sb strings.Builder

	sb.WriteString("\n  ")
	for i, cat := range categories {
		label := strings.ToUpper(cat[:1]) + cat[1:]
		if i == m.catIndex {
			sb.WriteString(common.StyleActiveTab.Render(label))
		} else {
			sb.WriteString(common.StyleInactiveTab.Render(label))
		}
	}
	sb.WriteString("\n")

	if m.loading {
		sb.WriteString("\n  " + m.spinner.View() + " Loading marketplace...\n")
		return sb.String()
	}

	if m.err != nil {
		sb.WriteString("\n  " + common.StyleError.Render("Error: "+m.err.Error()) + "\n")
		return sb.String()
	}

	sb.WriteString(m.list.View())
	sb.WriteString("\n" + common.StyleDim.Render("  ←/→ category  /filter  enter select"))

	return sb.String()
}

func (m *Marketplace) ClearSelection() {
	m.SelectedItem = nil
}

func (m *Marketplace) applyFilter() {
	cat := categories[m.catIndex]
	var items []list.Item
	for _, p := range m.allItems {
		if cat != "all" && p.Category != cat {
			continue
		}
		items = append(items, marketplaceItem{item: p})
	}
	m.list.SetItems(items)
}

func fetchMarketplace() tea.Msg {
	client := registry.NewClient()
	mp, err := client.FetchMarketplace()
	if err != nil {
		return common.MarketplaceFetchedMsg{Err: err}
	}
	return common.MarketplaceFetchedMsg{Marketplace: mp}
}
