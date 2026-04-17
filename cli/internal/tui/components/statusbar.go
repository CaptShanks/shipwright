package components

import (
	"strings"

	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
	"github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
	ViewName string
	HintKeys string
	Version  string
	Width    int
}

func (s StatusBar) View() string {
	left := common.StyleStatusBar.Render(" " + s.ViewName + " ")
	right := common.StyleStatusBar.Render(" ship " + s.Version + " ")

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)

	gap := s.Width - leftW - rightW - lipgloss.Width(s.HintKeys)
	if gap < 0 {
		gap = 0
	}

	mid := common.StyleStatusBar.Render(s.HintKeys + strings.Repeat(" ", gap))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, mid, right)
}
