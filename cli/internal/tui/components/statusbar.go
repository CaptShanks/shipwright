package components

import (
	"strings"

	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
	"github.com/charmbracelet/lipgloss"
)

var (
	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(common.ColorDim).
			Padding(0, 1)

	descStyle = lipgloss.NewStyle().
			Foreground(common.ColorDim)

	barStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236"))

	versionStyle = lipgloss.NewStyle().
			Foreground(common.ColorPrimary).
			Bold(true)

	viewLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(common.ColorPrimary).
			Padding(0, 1)
)

type KeyHint struct {
	Key  string
	Desc string
}

type StatusBar struct {
	ViewName string
	Hints    []KeyHint
	Version  string
	Width    int
}

func renderHint(h KeyHint) string {
	return keyStyle.Render(h.Key) + " " + descStyle.Render(h.Desc)
}

func (s StatusBar) View() string {
	var parts []string
	for _, h := range s.Hints {
		parts = append(parts, renderHint(h))
	}
	hintsLine := strings.Join(parts, "   ")

	label := viewLabelStyle.Render(s.ViewName)
	ver := versionStyle.Render("ship " + s.Version)

	topGap := s.Width - lipgloss.Width(label) - lipgloss.Width(ver) - 2
	if topGap < 0 {
		topGap = 0
	}
	topLine := barStyle.Width(s.Width).Render(
		label + strings.Repeat(" ", topGap) + ver,
	)

	hintW := lipgloss.Width(hintsLine)
	pad := (s.Width - hintW) / 2
	if pad < 0 {
		pad = 0
	}
	bottomLine := barStyle.Width(s.Width).Render(
		strings.Repeat(" ", pad) + hintsLine,
	)

	return topLine + "\n" + bottomLine
}
