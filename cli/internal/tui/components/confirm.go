package components

import (
	"github.com/CaptShanks/shipwright/cli/internal/tui/common"
	"github.com/charmbracelet/lipgloss"
)

type ConfirmDialog struct {
	Title   string
	Message string
	Visible bool
}

func (c ConfirmDialog) View(width, height int) string {
	if !c.Visible {
		return ""
	}

	body := common.StyleBold.Render(c.Title) + "\n\n" +
		c.Message + "\n\n" +
		common.StyleDim.Render("[y] confirm  [n/esc] cancel")

	box := common.StyleOverlay.Width(40).Render(body)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
