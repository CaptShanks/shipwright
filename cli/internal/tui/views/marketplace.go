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

type marketplaceDelegate struct {
	selected map[string]bool
}

func (d marketplaceDelegate) Height() int                             { return 3 }
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
	catColor := common.CategoryColor(cat)

	nameStyle := lipgloss.NewStyle().Bold(true)
	if selected {
		nameStyle = nameStyle.Foreground(catColor)
	}

	icon := common.CategoryIcons[cat]
	badge := common.CategoryBadge(cat)
	verStr := common.StyleDim.Render("v" + ver)

	checked := d.selected[name]
	checkBox := lipgloss.NewStyle().Foreground(common.ColorDim).Render("○")
	if checked {
		checkBox = lipgloss.NewStyle().Foreground(common.ColorSuccess).Bold(true).Render("●")
	}

	cursor := " "
	if selected {
		cursor = lipgloss.NewStyle().Foreground(catColor).Render("▸")
	}

	line1 := fmt.Sprintf(" %s %s %s %s  %s  %s", cursor, checkBox, icon, nameStyle.Render(name), badge, verStr)

	maxDesc := m.Width() - 8
	if maxDesc > 0 && len(desc) > maxDesc {
		desc = desc[:maxDesc-3] + "..."
	}
	line2 := common.StyleDim.Render("     " + desc)

	var tagLine string
	if len(mi.item.Tags) > 0 {
		var tags []string
		for _, t := range mi.item.Tags {
			tags = append(tags, lipgloss.NewStyle().Foreground(catColor).Render("#"+t))
		}
		tagLine = "     " + strings.Join(tags, " ")
	}

	border := lipgloss.NewStyle().Foreground(catColor)
	if selected {
		border = border.Bold(true)
	}
	separator := border.Render("     " + strings.Repeat("─", max(m.Width()-10, 20)))

	fmt.Fprintf(w, "%s\n%s\n%s%s", line1, line2, tagLine, separator)
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
	SelectedItem  *registry.MarketplaceItem
	selected      map[string]bool
	SelectedBatch []registry.MarketplaceItem
}

func NewMarketplace() Marketplace {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	sel := make(map[string]bool)

	l := list.New(nil, marketplaceDelegate{selected: sel}, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()
	l.KeyMap.Quit = key.NewBinding(key.WithDisabled())

	return Marketplace{
		list:     l,
		spinner:  sp,
		loading:  true,
		selected: sel,
	}
}

func (m Marketplace) SelectedCount() int {
	return len(m.selected)
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
			case " ":
				if item, ok := m.list.SelectedItem().(marketplaceItem); ok {
					name := item.item.Name
					if m.selected[name] {
						delete(m.selected, name)
					} else {
						m.selected[name] = true
					}
				}
				return m, nil
			case "a":
				if len(m.selected) == len(m.list.Items()) {
					for k := range m.selected {
						delete(m.selected, k)
					}
				} else {
					for _, li := range m.list.Items() {
						if mi, ok := li.(marketplaceItem); ok {
							m.selected[mi.item.Name] = true
						}
					}
				}
				return m, nil
			case "I":
				if len(m.selected) > 0 {
					var batch []registry.MarketplaceItem
					for _, p := range m.allItems {
						if m.selected[p.Name] {
							batch = append(batch, p)
						}
					}
					m.SelectedBatch = batch
					return m, nil
				}
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
		icon := ""
		if cat != "all" {
			icon = common.CategoryIcons[cat] + " "
		}
		label := strings.ToUpper(cat[:1]) + cat[1:]
		if i == m.catIndex {
			catColor := common.CategoryColor(cat)
			if cat == "all" {
				catColor = common.ColorPrimary
			}
			style := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("0")).
				Background(catColor).
				Padding(0, 2)
			sb.WriteString(style.Render(icon + label))
		} else {
			sb.WriteString(common.StyleInactiveTab.Render(icon + label))
		}
	}

	count := len(m.list.Items())
	sb.WriteString("  " + common.StyleDim.Render(fmt.Sprintf("(%d)", count)))
	sb.WriteString("\n")

	if selCount := len(m.selected); selCount > 0 {
		selBadge := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(common.ColorSuccess).
			Padding(0, 1).
			Render(fmt.Sprintf("✓ %d selected", selCount))
		sb.WriteString("  " + selBadge + "  ")
		sb.WriteString(common.StyleDim.Render("press ") +
			lipgloss.NewStyle().Bold(true).Foreground(common.ColorSuccess).Render("I") +
			common.StyleDim.Render(" to batch install"))
		sb.WriteString("\n")
	}

	if m.loading {
		sb.WriteString("\n  " + m.spinner.View() + " Loading marketplace...\n")
		return sb.String()
	}

	if m.err != nil {
		sb.WriteString("\n  " + common.StyleError.Render("Error: "+m.err.Error()) + "\n")
		return sb.String()
	}

	sb.WriteString(m.list.View())

	return sb.String()
}

func (m *Marketplace) ClearSelection() {
	m.SelectedItem = nil
}

func (m *Marketplace) ClearBatch() {
	m.SelectedBatch = nil
	for k := range m.selected {
		delete(m.selected, k)
	}
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
