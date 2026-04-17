package common

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("12")
	ColorSuccess   = lipgloss.Color("10")
	ColorError     = lipgloss.Color("9")
	ColorWarn      = lipgloss.Color("11")
	ColorDim       = lipgloss.Color("8")
	ColorHighlight = lipgloss.Color("14")
	ColorAccent    = lipgloss.Color("13")

	StyleBold   = lipgloss.NewStyle().Bold(true)
	StyleDim    = lipgloss.NewStyle().Foreground(ColorDim)
	StyleError  = lipgloss.NewStyle().Foreground(ColorError)
	StyleWarn   = lipgloss.NewStyle().Foreground(ColorWarn)
	StyleOK     = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleAccent = lipgloss.NewStyle().Foreground(ColorAccent)

	StyleActiveTab = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(ColorPrimary).
			Padding(0, 2)

	StyleInactiveTab = lipgloss.NewStyle().
				Foreground(ColorDim).
				Padding(0, 2)

	StyleStatusBar = lipgloss.NewStyle().
			Foreground(ColorDim).
			Background(lipgloss.Color("236"))

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDim).
			Padding(1, 2)

	StyleCardActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(1, 2)

	StyleOverlay = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	StyleBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(ColorSuccess).
			Padding(0, 1)

	StyleBadgeWarn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(ColorWarn).
			Padding(0, 1)

	ColorAgent  = lipgloss.Color("33")
	ColorSkill  = lipgloss.Color("178")
	ColorMCP    = lipgloss.Color("36")
	ColorBundle = lipgloss.Color("135")

	StyleBadgeAgent = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(ColorAgent).
				Bold(true).
				Padding(0, 1)

	StyleBadgeSkill = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(ColorSkill).
				Bold(true).
				Padding(0, 1)

	StyleBadgeMCP = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(ColorMCP).
			Bold(true).
			Padding(0, 1)

	StyleBadgeBundle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(ColorBundle).
				Bold(true).
				Padding(0, 1)
)

var CategoryIcons = map[string]string{
	"agents":  "🤖",
	"skills":  "⚡",
	"mcps":    "🔌",
	"bundles": "📦",
}

func CategoryBadge(category string) string {
	switch category {
	case "agents":
		return StyleBadgeAgent.Render("AGENT")
	case "skills":
		return StyleBadgeSkill.Render("SKILL")
	case "mcps":
		return StyleBadgeMCP.Render("MCP")
	case "bundles":
		return StyleBadgeBundle.Render("BUNDLE")
	default:
		return StyleDim.Render(category)
	}
}

func CategoryColor(category string) lipgloss.Color {
	switch category {
	case "agents":
		return ColorAgent
	case "skills":
		return ColorSkill
	case "mcps":
		return ColorMCP
	case "bundles":
		return ColorBundle
	default:
		return ColorDim
	}
}
