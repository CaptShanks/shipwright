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
)
